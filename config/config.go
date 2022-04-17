// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package config sets up the configuration for the server. The configuration
// settings may be from the default built-in configuration, read from a file,
// read from an io.Reader or specifically set.
//
// Typically the default built-in configuration is created using config.Default
// and then settings from a configuration file applied using config.Load, as
// this allows for omitted values in the configuration file to use the
// defaults. The configuration is not used directly, instead it should be
// passed to the various package level Config methods by the main function.
//
// The default server configuration is documented in docs/configuration-file.txt
package config

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/text"
)

// DefaultCfg is the default built-in server configuration.
const DefaultCfg = `// Built-in configuration
		Server.Host:          127.0.0.1
		Server.Port:          4001
		Server.IdleTimeout:   10m
		Server.MaxPlayers:    1024
		Stats.Rate:           10s
		Inventory.CrowdSize:  11
		Login.AccountLength:  10
		Login.PasswordLength: 10
		Login.SaltLength:     32
		Login.Timeout:        1m


WolfMUD Copyright 1984-2021 Andrew 'Diddymus' Rolfe

    [GREEN]W[WHITE]orld
    [GREEN]O[WHITE]f
    [GREEN]L[WHITE]iving
    [GREEN]F[WHITE]antasy

[YELLOW]Welcome to WolfMUD![RESET]

%%`

// Config represents the server's configuration.
type Config struct {
	Server    Server
	Quota     Quota
	Stats     Stats
	Inventory Inventory
	Login     Login
	Debug     Debug
	Greeting  string
}

type Server struct {
	Host        string
	Port        string
	IdleTimeout time.Duration
	MaxPlayers  int
	LogClient   bool
	DataPath    string // Calculated data path without trailing separator
}

type Quota struct {
	Slots  int
	Window time.Duration
	Stats  int
}

type Stats struct {
	Rate time.Duration
	GC   bool
}

type Inventory struct {
	CrowdSize int
}

type Login struct {
	AccountLength  int
	PasswordLength int
	SaltLength     int
	Timeout        time.Duration
}

type Debug struct {
	LongLog bool
	Panic   bool
	Events  bool
	Things  bool
	Quota   bool
}

// Default returns the default, built-in server configuration. Failure to
// create the default built-in configuration is fatal.
func Default() Config {
	var (
		c   Config
		err error
	)

	c, err = c.Read(bytes.NewBufferString(DefaultCfg))
	if err != nil {
		log.Fatalf("Cannot create built-in default configuration: %s", err)
	}
	c.Server.DataPath = filepath.Join("..", "data")
	return c
}

// Load loads the configuration specified by the passed path. The passed path
// may be relative or absolute, a path ending in a path separator (in which
// case the default file 'config.wrj' will be used) or a configuration file
// name (which will be prefixed with the default path e.g. '../data').
//
// The receiver is used as a base configuration and the loaded configuration is
// applied over it. This allows settings omitted from the loaded configuration
// to use any defaults set in the base configuration. Returns the new
// configuration and a nil error on success. On failure returns a copy of the
// base configuration and a non-nil error.
func (c Config) Load(path string) (Config, error) {
	path = c.resolvePath(path)
	log.Printf("Using configuration file: %s", path)
	r, err := os.Open(path)
	if err != nil {
		return c, err
	}
	defer r.Close()
	c.Server.DataPath = filepath.Dir(path)
	return c.Read(r)
}

// Read reads configuration settings from the passed io.Reader. The receiver is
// used as a base configuration and the read settings are applied over it. This
// allows settings omitted from the read configuration to use any defaults set
// in the base configuration. Returns the new configuration and nil error on
// success. On failure returns a copy of the base configuration and a non-nil
// error.
func (c Config) Read(r io.Reader) (Config, error) {
	jar := recordjar.Read(r, "GREETING")
	for _, rec := range jar {
		for field, data := range rec {
			switch field {

			// Server settings
			case "SERVER.HOST":
				c.Server.Host = decode.String(data)
			case "SERVER.PORT":
				c.Server.Port = decode.String(data)
			case "SERVER.IDLETIMEOUT":
				c.Server.IdleTimeout = decode.Duration(data)
			case "SERVER.MAXPLAYERS":
				c.Server.MaxPlayers = decode.Integer(data)
			case "SERVER.LOGCLIENT":
				c.Server.LogClient = decode.Boolean(data)

			// Quota settings
			case "QUOTA.SLOTS":
				c.Quota.Slots = decode.Integer(data)
			case "QUOTA.WINDOW":
				c.Quota.Window = decode.Duration(data)
			case "QUOTA.STATS":
				c.Quota.Stats = decode.Integer(data)

			// Stats settings
			case "STATS.RATE":
				c.Stats.Rate = decode.Duration(data)
			case "STATS.GC":
				c.Stats.GC = decode.Boolean(data)

			// Inventory settings
			case "INVENTORY.CROWDSIZE":
				c.Inventory.CrowdSize = decode.Integer(data)

			// Login settings
			case "LOGIN.ACCOUNTLENGTH":
				c.Login.AccountLength = decode.Integer(data)
			case "LOGIN.PASSWORDLENGTH":
				c.Login.PasswordLength = decode.Integer(data)
			case "LOGIN.SALTLENGTH":
				c.Login.SaltLength = decode.Integer(data)
			case "LOGIN.TIMEOUT":
				c.Login.Timeout = decode.Duration(data)

			// Debug settings
			case "DEBUG.LONGLOG":
				c.Debug.LongLog = decode.Boolean(data)
			case "DEBUG.PANIC":
				c.Debug.Panic = decode.Boolean(data)
			case "DEBUG.EVENTS":
				c.Debug.Events = decode.Boolean(data)
			case "DEBUG.THINGS":
				c.Debug.Things = decode.Boolean(data)
			case "DEBUG.QUOTA":
				c.Debug.Quota = decode.Boolean(data)

			case "GREETING":
				c.Greeting = string(text.Colorize(data))

			}
		}
	}
	return c, nil
}

// resolvePath tries to locate the configuration file to use from the passed
// path. If path is the empty string the default is used. If path looks like
// plain file name, without any directories, it is prefixed with the default
// directory. If path looks like plain directories, without a trailing file
// name, the default config file is appended. The resulting configuration file
// path to use is then returned.
func (c Config) resolvePath(path string) string {

	if path == "" {
		return filepath.Join("..", "data", "config.wrj")
	}

	dir, file := filepath.Split(path)

	if dir == "" {
		dir = filepath.Join("..", "data")
	}
	if file == "" {
		file = "config.wrj"
	}
	return filepath.Join(dir, file)
}
