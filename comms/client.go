// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms

import (
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/frontend"
	"code.wolfmud.org/WolfMUD.git/text"

	"bufio"
	"io"
	"log"
	"net"
	"time"
)

// TODO: These need to be configuration options once we have them
const (
	termColumns = 80
	termLines   = 24
)

// This interface lets us assert network or our own errors
type temporary interface {
	Temporary() bool
}

// client contains state information about a client connection. The err field
// should not be manipulated directly. Instead call Error() and SetError().
//
// The current frontend in use is an anonymous interface as this lets us define
// what type frontend is - even though we don't have access to the unexported
// frontend struct from the frontend package.
//
// TODO: client is currently talking directly to a player. It should be talking
// to a switchable, abstract layer so that we can talk to a player, menus,
// account system etc.
type client struct {
	*net.TCPConn            // The client's network connection
	remoteAddr   string     // Client's remote address
	err          chan error // Error channel to sync between input & output

	frontend interface { // The current frontend in use
		Parse([]byte) error
		Close()
	}
}

// newClient returns an initialised client for the passed connection.
func newClient(conn *net.TCPConn) *client {

	// Setup connection parameters
	conn.SetKeepAlive(true)
	conn.SetLinger(10)
	conn.SetNoDelay(false)
	conn.SetWriteBuffer(termColumns * termLines)
	conn.SetReadBuffer(termColumns)

	c := &client{
		TCPConn:    conn,
		remoteAddr: conn.RemoteAddr().String(),
		err:        make(chan error, 1),
	}

	c.err <- nil

	c.frontend = frontend.New(c)
	c.frontend.Parse([]byte(""))

	return c
}

// process handles input from the network connection.
func (c *client) process() {

	// Main input processing loop, terminates on any error raised not just read
	// or Parse errors.
	{
		// Variables for use in the loop only hence the scoping outer braces
		var (
			s   = bufio.NewReaderSize(c, termColumns) // Sized network read buffer
			err error                                 // function local errors
			in  []byte                                // Input string from buffer
		)

		for c.Error() == nil {
			c.SetReadDeadline(time.Now().Add(config.Server.IdleTimeout))
			if in, err = s.ReadBytes('\n'); err != nil {
				c.SetError(err)
				continue
			}
			if err = c.frontend.Parse(in); err != nil {
				c.SetError(err)
			}
		}
	}

	// Deallocate current frontend
	c.frontend.Close()
	c.frontend = nil

	// If connection timed out notify the client
	if oe, ok := c.Error().(*net.OpError); ok && oe.Timeout() {
		c.Write([]byte("\n\nIdle connection terminated by server.\n"))
	}

	// Say goodbye to client
	c.Write([]byte("\nBye bye...\n\n"))

	// io.EOF does not give address info so handle specially, otherwise just
	// report the error
	if c.Error() == io.EOF {
		log.Printf("Connection dropped by remote client: %s", c.remoteAddr)
	} else {
		log.Printf("Connection error: %s, %s", c.Error(), c.remoteAddr)
	}

	// Make sure connection closed down and deallocated
	if err := c.Close(); err != nil {
		log.Printf("Error closing connection: %s", err)
	} else {
		log.Printf("Connection closed: %s", c.remoteAddr)
	}
	c.TCPConn = nil

	// Close and drain error channel
	close(c.err)
	<-c.err
}

// Write handles output for the network connection.
func (c *client) Write(d []byte) (n int, err error) {

	// If we already have a non-temporary error do nothing
	if e := c.Error(); e != nil {
		if e, ok := e.(temporary); !ok || !e.Temporary() {
			return
		}
	}

	var t []byte

	if len(d) != 0 {
		t = text.Fold(d, termColumns)
	}

	c.SetWriteDeadline(time.Now().Add(config.Server.IdleTimeout))

	if n, err = c.TCPConn.Write(t); err != nil {
		c.SetError(err)
	}
	return
}

// Error returns the first error raised or nil if there is no error. An error
// can be set by calling SetError.
func (c *client) Error() (err error) {
	err = <-c.err
	c.err <- err
	return err
}

// SetError is used to record the first error condition that occurs. Subsequent
// calls will not over write the initial error raised. The current error can be
// checked by calling Error.
func (c *client) SetError(err error) {
	e := <-c.err
	if e == nil {
		e = err
	}
	c.err <- e
}
