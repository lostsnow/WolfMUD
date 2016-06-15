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

// Values to be treated as constants but we can't define them as constants
var (
	defaultPrompt = []byte(">") // Default prompt for client input
	noPrompt      = []byte{}    // An empty prompt
)

// This interface lets us assert network or our own errors
type temporary interface {
	Temporary() bool
}

// client contains state information about a client connection. The err field
// should not be manipulated directly. Instead call Error() and SetError().
//
// TODO: client is currently talking directly to a player. It should be talking
// to a switchable, abstract layer so that we can talk to a player, menus,
// account system etc.
type client struct {
	*net.TCPConn                  // The client's network connection
	driver       *frontend.Driver // The current driver in use
	remoteAddr   string           // Client's remote address
	err          chan error       // Error channel to sync between input & output
	prompt       []byte           // The current prompt the client is using
}

// newClient returns an initialised client for the passed connection.
func newClient(conn *net.TCPConn) *client {

	// Setup connection parameters
	conn.SetKeepAlive(true)
	conn.SetLinger(0)
	conn.SetNoDelay(false)
	conn.SetWriteBuffer(termColumns * termLines)
	conn.SetReadBuffer(termColumns)

	c := &client{
		TCPConn:    conn,
		remoteAddr: conn.RemoteAddr().String(),
		err:        make(chan error, 1),
		prompt:     defaultPrompt,
	}

	c.err <- nil

	c.driver = frontend.NewDriver(c)
	c.driver.Parse([]byte(""))

	return c
}

// process handles input from the network connection.
func (c *client) process() {

	var (
		s   = bufio.NewReaderSize(c, termColumns) // Sized network read buffer
		err error                                 // function local errors
		in  []byte                                // Input string from buffer
	)

	// Main input processing loop, terminates on any error raised not just read
	// or Parse errors.
	for c.Error() == nil {
		c.SetReadDeadline(time.Now().Add(config.Server.IdleTimeout))
		if in, err = s.ReadBytes('\n'); err != nil {
			c.SetError(err)
			continue
		}
		if err = c.driver.Parse(in); err != nil {
			c.SetError(err)
		}
	}

	// If connection time out clear timeout error to notify the client
	if oe, ok := err.(*net.OpError); ok && oe.Timeout() {
		<-c.err
		c.err <- nil
		c.prompt = noPrompt
		c.Write([]byte("\n\nIdle connection terminated by server.\n"))
	}

	// If error is temporary clear error and say goodbye to client
	if oe, ok := err.(temporary); ok && oe.Temporary() {
		<-c.err
		c.err <- nil
		c.prompt = noPrompt
		c.Write([]byte("\nBye bye...\n\n"))
	}

	// io.EOF does not give address info so handle specially, otherwise just
	// report the error
	if err == io.EOF {
		log.Printf("Connection dropped by remote client: %s", c.remoteAddr)
	} else {
		log.Printf("Connection error: %s", err)
	}

	// Deallocate current driver
	c.driver.Close()
	c.driver = nil

	// Make sure connection closed down and deallocated
	if err = c.Close(); err != nil {
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

	// Don't try doing anything if we already have errors
	if c.Error() != nil {
		return
	}

	var t []byte

	if len(d) != 0 {
		t = text.Fold(d, termColumns)
		t = append(t, "\r\n"...)
		t = append(t, c.prompt...)
	} else {
		t = c.prompt
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
