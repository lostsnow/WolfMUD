// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms

import (
	"log"
	"net"

	"code.wolfmud.org/WolfMUD.git/text"
)

// Message sent to client when they have been banned
const tooManyMsg = text.Bad +
	"\nToo many repeat connections. Please try again later.\n\n" + text.Reset

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

	seq := uint64(0)
	q := NewQuota()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
			continue
		}

		// Check if IP address is over its quota. If it is close the connection.
		// Note that IP addresses that cannot be parsed will share a common quota.
		if q.Enabled() {
			ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
			if q.Quota(ip) {
				conn.CloseRead()
				conn.Write([]byte(tooManyMsg))
				conn.SetKeepAlive(false)
				conn.SetLinger(0)
				conn.CloseWrite()
				continue
			}
		}

		c := newClient(conn, seq)
		go c.process()
		seq++
	}
}
