// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package world holds references to all of the locations in the world and
// accepts new client connections.
package world

import (
	crypto "crypto/rand"
	"log"
	"math/big"
	"math/rand"
	"net"
	"wolfmud.org/client"
	"wolfmud.org/entities/location"
	"wolfmud.org/entities/location/startingLocation"
	"wolfmud.org/utils/stats"
)

const (
	HOST = "127.0.0.1" // Host to listen on
	PORT = "4001"      // Port to listen on
)

// World represents a single game world. It has references to all of the
// locations available in it. The locations could be held in an inventory but
// that is overkill in this situation so we use a slice of locations.
type World struct {
	locations         []location.Interface
	startingLocations []*startingLocation.StartingLocation
}

// New creates a new World and returns a reference to it.
func New() *World {

	// Initialise random number generator with random seed
	seed, _ := crypto.Int(crypto.Reader, big.NewInt(0x7FFFFFFFFFFFFFFF))
	rand.Seed(seed.Int64())

	return &World{}
}

// Genesis starts the world - what else? :) Genesis opens the listening server
// socket and accepts connections. It also starts the stats Goroutine.
func (w *World) Genesis() {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting WolfMUD server...")

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(HOST, PORT))
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
			go client.Spawn(conn, w.startingLocations[rand.Intn(len(w.startingLocations))])
		}
	}
}

// AddLocation adds a location to the list of locations for this world.
func (w *World) AddLocation(l location.Interface) {
	w.locations = append(w.locations, l)

	if sl, ok := l.(*startingLocation.StartingLocation); ok {
		w.startingLocations = append(w.startingLocations, sl)
	}
}
