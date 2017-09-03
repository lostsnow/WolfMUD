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
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Server default configuration
var Server = struct {
	Host           string        // Host for server to listen on
	Port           string        // Port for server to listen on
	Greeting       []byte        // Connection greeting
	IdleTimeout    time.Duration // Idle connection disconnect time
	MaxPlayers     int           // Max number of players allowed to login at once
	DataDir        string        // Main data directory
	SetPermissions bool          // Set permissions on created account files?
}{
	Host:           "127.0.0.1",
	Port:           "4001",
	Greeting:       []byte(""),
	IdleTimeout:    10 * time.Minute,
	MaxPlayers:     1024,
	DataDir:        ".",
	SetPermissions: false,
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

// Debugging configuration
var Debug = struct {
	LongLog    bool // Long log with microseconds & filename?
	Panic      bool // Let goroutines panic and stop server?
	AllowDump  bool // Allow use of #DUMP command?
	AllowDebug bool // Allow use of #DEBUG command?
	Events     bool // Log events? - this can make the log quite noisy
	Things     bool // Log additional information for Thing?
}{
	LongLog:    false,
	Panic:      false,
	AllowDump:  false,
	AllowDebug: false,
	Events:     false,
	Things:     false,
}

// Load loads the configuration file and overrides the default configuration
// values with any values found.
func init() {

	// Setup global logging format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)

	// Seed default random source
	rand.Seed(time.Now().UnixNano())

	f, err := openConfig()

	if err != nil {
		log.Printf("Configuration file error: %s", err)
		return
	}

	if f == nil {
		log.Print("No configuration file used. Using defaults.")
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
		case "SERVER.MAXPLAYERS":
			Server.MaxPlayers = recordjar.Decode.Integer(data)
		case "SERVER.GREETING":
			Server.Greeting = text.Colorize(recordjar.Decode.Bytes(data))

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

		// Debug settings
		case "DEBUG.LONGLOG":
			Debug.LongLog = recordjar.Decode.Boolean(data)
		case "DEBUG.PANIC":
			Debug.Panic = recordjar.Decode.Boolean(data)
		case "DEBUG.ALLOWDUMP":
			Debug.AllowDump = recordjar.Decode.Boolean(data)
		case "DEBUG.ALLOWDEBUG":
			Debug.AllowDebug = recordjar.Decode.Boolean(data)
		case "DEBUG.EVENTS":
			Debug.Events = recordjar.Decode.Boolean(data)
		case "DEBUG.THINGS":
			Debug.Things = recordjar.Decode.Boolean(data)

		// Unknow setting
		default:
			log.Printf("Unknown setting %s: %s", field, data)
		}
	}

	log.Printf("Data Path: %s", Server.DataDir)

	Server.SetPermissions, err = filesystemCheck(Server.DataDir)
	log.Printf("Set permissions on player account files: %t", Server.SetPermissions)
	if err != nil {
		log.Printf("Error checking permissions, %s", err)
	}

	if !Debug.LongLog {
		log.SetFlags(log.Ldate | log.Ltime)
		log.Printf("Switching to short log format.")
	}
}

// openConfig tries to locate and open the configuration file to use. By
// default it will use the path specified on the command line. As a fallback it
// will use the data directory in the current directory. If the path does not
// specify a filename the default config.wrj will be used. If a configuration
// file is found and accessible a *os.File to it will be returned with a nil
// error. If not found a nil pointer and a non-nil error will be returned.
func openConfig() (config *os.File, err error) {

	// Has user supplied path ± specific file?
	flag.Parse()
	dir, file := filepath.Split(flag.Arg(0))

	if dir == "" && strings.ToUpper(file) == "NONE" {
		return nil, nil
	}

	// Is the file actually a directory without a final separator?
	if file != "" && filepath.Ext(file) != ".wrj" {
		dir = filepath.Join(dir, file)
		file = ""
	}

	// If no user supplied path use the data directory in the current working
	// directory
	if dir == "" {
		if dir, err = os.Getwd(); err != nil {
			return nil, err
		}
		dir = filepath.Join(dir, "data")
	}

	// If no configuration file provided use the default
	if file == "" {
		file = "config.wrj"
	}

	// Make sure path + file is good
	path := filepath.Join(dir, file)
	if path, err = filepath.Abs(path); err != nil {
		return nil, err
	}

	// Try and open configuration file
	if config, err = os.Open(path); err != nil {
		return nil, err
	}

	log.Printf("Found configuration file: %s", path)
	return config, nil
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

// filesystemCheck tests to see if the filesystem the passed path is on
// supports the changing of file permissions. If they do true will be returned,
// otherwise false. The returned error will be non-nil if an error occurs when
// checking.
func filesystemCheck(path string) (bool, error) {

	p := filepath.Join(path, "check.tmp")

	defer os.Remove(p)

	var (
		f    *os.File
		info os.FileInfo
		err  error
	)

	if err = os.Remove(p); err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
	}

	if f, err = os.Create(p); err != nil {
		return false, err
	}

	if err = f.Chmod(0660); err != nil {
		return false, err
	}

	if info, err = os.Stat(p); err != nil {
		return false, err
	}

	if err = f.Close(); err != nil {
		return false, err
	}

	if info.Mode()&os.FileMode(0777) != os.FileMode(0660) {
		return false, fmt.Errorf("Cannot set permissions to 0660: 0%o", info.Mode())
	}

	return true, nil
}
