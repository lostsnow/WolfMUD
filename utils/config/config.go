// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package config centralizes all of the tweakable configuration settings.
// Initially sensible default values are configured but can be overridden with
// a config.wrj file. This will be looked for in a number of locations:
//
//	- The current directory
//	- A data directory in the current directory
//	- The directory of the server binary
//	- A data directory in the binary directory
//	- In the repository data directory:
//				bin/../src/code.wolfmud.org/WolfMUD.git/data
//
// This should allow for the binary and data files to be downloaded and placed
// in the same directory, for the data to be put into a separate data directory
// with the binary, for the data to be put anywhere and used as the current
// directory. Lastly it also allows for the Git repository to be cloned and a
// standard Go workspace to be used.
//
// TODO: Allow configuration to be reloaded for a running server. This means
// that all configuration values will need to be accessed directly each time
// they are used and not cached.
package config

import (
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// TODO: configName should be overrideable by a command line option
var (
	searchPaths = []string{}
	configName  = "config.wrj"
)

// All configuration values with initial sensible defaults. This allows the
// values to be accessed simplay as config.<property>
var (
	DataDir                = "."
	ListenAddress          = "127.0.0.1"
	ListenPort             = "4001"
	MemProfileRate     int = 0
	StatsRate              = 5 * time.Minute
	AccountIdMin           = 10
	AccountPasswordMin     = 10
)

// init sets up the configuration search paths:
//
//	1st = cwd
//	2nd = cwd/data
//	3rd = bin
//	4th = bin/data
//	5th = bin/../src/code.wolfmud.org/WolfMUD.git/data
//
func init() {

	if cwd, err := os.Getwd(); err == nil {
		searchPaths = append(searchPaths, cwd)
		searchPaths = append(searchPaths, cwd+"/data")
	}

	if selfBin, err := filepath.Abs(filepath.Dir(os.Args[0])); err == nil {
		if len(searchPaths) == 0 || selfBin != searchPaths[0] {
			searchPaths = append(searchPaths, selfBin)
		}
		searchPaths = append(searchPaths, selfBin+"/data")
		searchPaths = append(searchPaths, selfBin+"/../data")
		searchPaths = append(searchPaths, selfBin+"/../src/code.wolfmud.org/WolfMUD.git/data")
	}

	// Make sure paths are in native format for OS
	for i, path := range searchPaths {
		searchPaths[i] = filepath.FromSlash(path)
	}

}

// BUG(Diddymus): A missing configuration value currently causes the default
// for the type to be assigned ignoring the sensible defaults.

// Read reads the config.wrj and sets new configuration values found in it.
//
// TODO: We should be able to reload and change settings at any time.
//
// TODO: Need to add more error checking of values coming in.
func Read() {

	ps := string(os.PathSeparator)

	for _, path := range searchPaths {
		log.Printf("Checking for %s in: %s", configName, path)
		if dir, err := os.Open(path + ps + configName); err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Error checking for %s: %s", configName, err)
				continue
			}
		} else {
			defer dir.Close()
			log.Printf("Using: %s%s%s", path, ps, configName)
			rj, _ := recordjar.Read(dir)
			d := recordjar.Decoder(rj[0])

			ListenAddress = d.String("listen.address")
			ListenPort = d.String("listen.port")
			runtime.MemProfileRate = d.Int("mem.profile.rate")
			StatsRate = d.Duration("stats.rate")
			AccountIdMin = d.Int("account.id.min")
			AccountPasswordMin = d.Int("account.password.min")

			DataDir, _ = filepath.Abs(path + ps + d["data.dir"])
			DataDir += ps

			log.Printf("listen.address: %s", ListenAddress)
			log.Printf("listen.port: %s", ListenPort)
			log.Printf("mem.profile.rate: %d", MemProfileRate)
			log.Printf("stats.rate: %s", StatsRate)
			log.Printf("data.dir: %s", DataDir)
			log.Printf("account.id.min: %d", AccountIdMin)
			log.Printf("account.password.min: %d", AccountPasswordMin)

			break
		}
	}

}
