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
	"code.wolfmud.org/WolfMUD.git/utils/loader"
	"code.wolfmud.org/WolfMUD.git/utils/stats"

	_ "code.wolfmud.org/WolfMUD.git/entities/thing/item"

	"log"
)

func main() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting WolfMUD server...")

	config.Read()
	stats.Start()
	loader.Load(config.DataDir)

	world.New(config.ListenAddress, config.ListenPort).Genesis()

	log.Println("WolfMUD server ending")

}
