// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package world holds references to all of the locations in the world and
// accepts new client connections.
package world

import (
	"fmt"
	"log"
	"net"
	"wolfmud.org/client"
	"wolfmud.org/entities/location"
	"wolfmud.org/entities/mobile/player"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/stats"
)

// greeting is displayed when a new client connects.
//
// TODO: Soft code with rest of settings.
const (
	greeting = `

WolfMUD © 2012 Andrew 'Diddymus' Rolfe

    World
    Of
    Living
    Fantasy

`
)

// World represents a single game world. It has references to all of the
// locations available in it.
type World struct {
	locations []location.Interface
}

// Create brings a new world into existance and returns a reference to it.
func Create() *World {
	return &World{}
}

// Genesis starts the world - what else? :) Genesis opens the listening server
// socket and accepts connections. It also starts the stats Goroutine.
func (w *World) Genesis() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting WolfMUD server...")

	addr, err := net.ResolveTCPAddr("tcp", "localhost:4001")
	if err != nil {
		log.Printf("Error resolving TCP address, %s. Server will now exit.\n", err)
		return
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("Error setting up listener, %s. Server will now exit.\n", err)
		return
	}

	log.Printf("Accepting connections on: %s\n", addr)

	stats.Start()

	for {
		if conn, err := listener.AcceptTCP(); err != nil {
			log.Printf("Error accepting connection: %s. Server will now exit.\n", err)
			return
		} else {
			log.Printf("Connection from %s.\n", conn.RemoteAddr().String())
			go client.Spawn(conn, w)
		}
	}
}

// AddLocation adds a location to the list of locations for this world.
func (w *World) AddLocation(l location.Interface) {
	w.locations = append(w.locations, l)
}

func (w *World) AddThing(t thing.Interface) {
	id := 0 //t.UniqueId() % 15
	w.locations[id].Lock()
	w.locations[id].Add(t)
	w.locations[id].Unlock()
}

func (w *World) Broadcast(ommit []thing.Interface, format string, any ...interface{}) {
	msg := fmt.Sprintf("\n"+format, any...)

	for _, p := range player.PlayerList.List(ommit...) {
		p.Respond(msg)
	}
}
