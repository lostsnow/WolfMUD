// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// server is the main executable command used to start a WolfMUD server running.
// Currently it takes no parameters.
package main

import (
	"code.wolfmud.org/WolfMUD.git/entities/world"
	"code.wolfmud.org/WolfMUD.git/utils/config"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"
	"code.wolfmud.org/WolfMUD.git/utils/stats"

	_ "code.wolfmud.org/WolfMUD.git/entities/thing/item"

	"log"
)

// version is set at compile time by passing in e.g:
//
//		-ldflags "-X main.version $(git describe --dirty)"
//
// this will set version to something like:
//
//		PROTOTYPE1-110-g233321a
//
// If there are uncomitted changes '-dirty' will be appended to the end of the
// version string. This is a very handy reference when debugging other people
// issues.
//
// Not sure how this will pan out so will have to wait and see...
var version string

func main() {

	if version == "" {
		version = "Unknown Build"
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting WolfMUD server...")
	log.Printf("WolfMUD version: %s", version)

	config.Read()
	stats.Start()

	recordjar.LoadDir(config.DataDir)

	world.New(config.ListenAddress, config.ListenPort).Genesis()

	log.Println("WolfMUD server ending")

}
