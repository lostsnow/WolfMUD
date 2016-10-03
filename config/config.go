// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package config provides access to all of the tunable settings of a WolfMUD
// server. The default values can be overidden via a configuration file. The
// default name of the configuration file is config.wrj.
//
// Users may specify an alternate path for the configuration file on the
// command line. As a fallback it will use the current directory. If the path
// does not specify a filename the default config.wrj will be used.
package config

import (
	"code.wolfmud.org/WolfMUD.git/recordjar"

	"flag"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Server default configuration
var Server = struct {
	Host        string        // Host for server to listen on
	Port        string        // Port for server to listen on
	Greeting    []byte        // Connection greeting
	IdleTimeout time.Duration // Idle connection disconnect time
	DataDir     string        // Main data directory
}{
	Host:        "127.0.0.1",
	Port:        "4001",
	Greeting:    []byte(""),
	IdleTimeout: 10 * time.Minute,
	DataDir:     ".",
}

// Stats default configuration
var Stats = struct {
	Rate time.Duration // Stats collection and display rate
	GC   bool          // Run garbage collection before stat collection
}{
	Rate: 10 * time.Second,
	GC:   false,
}

// Inventory default configuration
var Inventory = struct {
	Compact   int // only compact if cap - len*2 > compact
	CrowdSize int // If inventory has more player than this it's a crowd
}{
	Compact:   4,
	CrowdSize: 10,
}

// Login default configuration
var Login = struct {
	AccountLength  int
	PasswordLength int
	SaltLength     int
}{
	AccountLength:  10,
	PasswordLength: 10,
	SaltLength:     32,
}

// Load loads the configuration file and overrides the default configuration
// values with any values found.
func Load() {

	f, err := openConfig()

	if err != nil {
		log.Printf("Configuration file error: %s", err)
		return
	}

	Server.DataDir = filepath.Dir(f.Name())
	log.Printf("Loading: %s", f.Name())

	j := recordjar.Read(f, "server.greeting")
	f.Close()
	record := j[0]

	// NOTE: a recordjar will uppercase all fieldnames so we need to use
	// uppercase switch cases below.
	for field, data := range record {
		switch field {

		// Main server settings
		case "SERVER.HOST":
			Server.Host = recordjar.Decode.String(data)
		case "SERVER.PORT":
			Server.Port = recordjar.Decode.String(data)
		case "SERVER.IDLETIMEOUT":
			Server.IdleTimeout = recordjar.Decode.Duration(data)
		case "SERVER.GREETING":
			Server.Greeting = recordjar.Decode.Bytes(data)

		// Stats settings
		case "STATS.RATE":
			Stats.Rate = recordjar.Decode.Duration(data)
		case "STATS.GC":
			Stats.GC = recordjar.Decode.Boolean(data)

		// Inventory settings
		case "INVENTORY.COMPACT":
			Inventory.Compact = recordjar.Decode.Integer(data)
		case "INVENTORY.CROWDSIZE":
			Inventory.CrowdSize = recordjar.Decode.Integer(data)

		// Login settings
		case "LOGIN.ACCOUNTLENGTH":
			Login.AccountLength = recordjar.Decode.Integer(data)
		case "LOGIN.PASSWORDLENGTH":
			Login.PasswordLength = recordjar.Decode.Integer(data)
		case "LOGIN.SALTLENGTH":
			Login.SaltLength = recordjar.Decode.Integer(data)

		// Unknow setting
		default:
			log.Printf("Unknown setting %s: %s", field, data)
		}
	}

	log.Printf("Data Path: %s", Server.DataDir)
}

// openConfig tries to locate and open the configuration file to use. By
// default it will use the path specified on the command line. As a fallback it
// will use the current directory. If the path does not specify a filename the
// default config.wrj will be used. If a configuration file is found and
// accessible a *os.File to it will be returned with a nil error. If not found
// a nil pointer and a non-nil error will be returned.
func openConfig() (f *os.File, err error) {

	// Has user supplied path ± specific file?
	flag.Parse()
	d := flag.Arg(0)

	// If no user supplied path ± file use the current working directory
	if d == "" {
		if d, err = os.Getwd(); err != nil {
			return nil, err
		}
	}

	// Make sure path ± file is good
	if d, err = filepath.Abs(d); err != nil {
		return nil, err
	}

	// Try getting information on path ± file
	var info os.FileInfo
	if info, err = os.Stat(d); err != nil {
		return nil, err
	}

	// If we just have a path add default filename
	if info.IsDir() {
		d = filepath.Join(d, "config.wrj")
	}

	// Try and open configuration file
	if f, err = os.Open(d); err != nil {
		return nil, err
	}

	log.Printf("Found configuration file: %s", d)
	return f, nil
}

// findData tries to locate the data directory relative to the configuration
// file.
func findData(base, path string) (data string, err error) {

	data = filepath.Join(base, path)

	// Make sure path is good
	if data, err = filepath.Abs(data); err != nil {
		return "", err
	}

	// Try getting information on path
	var info os.FileInfo
	if info, err = os.Stat(data); err != nil {
		return "", err
	}

	// If the path isn't a directory remove the filename.
	if !info.IsDir() {
		data = filepath.Dir(data)
	}

	return data, nil
}
