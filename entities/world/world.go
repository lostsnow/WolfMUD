// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package world waits accepting new client connections. When a new connection
// is made it gets it's own Goroutine to be handled.
package world

import (
	"code.wolfmud.org/WolfMUD.git/client"
	"log"
	"net"
)

// World represents the game world.
type World struct {
	host string
	port string
}

// New creates a new World and returns a reference to it.
func New(host, port string) *World {
	return &World{host, port}
}

// Genesis starts the world - what else? :) Genesis opens the listening server
// socket and accepts connections. It also starts the stats Goroutine.
func (w *World) Genesis() {

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(w.host, w.port))
	if err != nil {
		log.Printf("Error resolving local address: %s", err)
		return
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Printf("Error setting up listener: %s", err)
		return
	}

	log.Printf("Accepting connections on: %s", addr)

	for {
		if conn, err := listener.AcceptTCP(); err != nil {
			log.Printf("Error accepting connection: %s", err)
			return
		} else {
			log.Printf("Connection from: %s", conn.RemoteAddr().String())
			go client.Spawn(conn)
		}
	}
}
