// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// server is the main executable command used to start a WolfMUD server running.
// Currently it takes no parameters.
//
// TODO: There is no technical reason why a server cannot have multiple world
// instances and run multiple games. Depending on whether we want clients to
// connect to a single port and choose a world or have a port per world and
// users choose where the client connects the server socket code may need to be
// relocated from the world to server package.
package main

import (
	"code.wolfmud.org/WolfMUD.git/entities/world"
	"code.wolfmud.org/WolfMUD.git/utils/loader"
	"code.wolfmud.org/WolfMUD.git/utils/stats"
	"log"
	"runtime"
)

func main() {

	runtime.MemProfileRate = int(0)

	world := world.New()
	loader.Load(world)
	world.Genesis()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting WolfMUD server...")

	stats.Start()

	log.Println("WolfMUD server ending")

}
