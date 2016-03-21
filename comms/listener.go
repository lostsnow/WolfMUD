// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms

import (
	"log"
	"net"
	"runtime"
)

// Listen sets up a socket to listen for client connections. When a client
// connects the connection made is passed to newClient to setup a client
// instance for housekeeping. client.Process is then launched as a new
// goroutine to handle the main I/O processing for the client.
//
// TODO: currently there is no way to shut the server down other than Ctrl-C
func Listen(host, port string) {

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(host, port))
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
			log.Printf("Connection from: %s", conn.RemoteAddr())
			c := newClient(conn)
			go c.process()
		}
		runtime.Gosched()
	}
}
