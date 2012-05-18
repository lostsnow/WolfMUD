// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// server is the main executable command used to start a WolfMUD server running.
// Currently it takes no parameters                                            .
//
// TODO: There is no technical reason why a server cannot have multiple world
// instances and run multiple games. Depending on whether we want clients to
// connect to a single port and choose a world or have a port per world and
// users choose where the client connects the server socket code may need to be
// relocated from the world to server package.
package main

import (
	"runtime"
	"wolfmud.org/utils/loader"
	"wolfmud.org/entities/world"
)

func main() {

	runtime.MemProfileRate = int(0)

	world := world.Create()
	loader.Load(world)
	world.Genesis()

}
