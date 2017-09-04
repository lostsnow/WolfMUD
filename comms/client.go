// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"runtime/debug"
	"time"

	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/frontend"
	"code.wolfmud.org/WolfMUD.git/text"
)

// TODO: These need to be configuration options once we have them
const (
	termColumns = 80
	termLines   = 24
	inputBuffer = 512
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
	conn.SetReadBuffer(inputBuffer)

	c := &client{
		TCPConn:    conn,
		remoteAddr: conn.RemoteAddr().String(),
		err:        make(chan error, 1),
	}

	c.err <- nil

	c.leaseAcquire()

	// Setup frontend if no error acquiring a lease
	if c.Error() == nil {
		c.frontend = frontend.New(c)
		c.frontend.Parse([]byte(""))
	}

	return c
}

// process handles input from the network connection.
func (c *client) process() {

	// If a client goroutine panics try not to bring down the whole server down
	// unless the configuration option Debug.Panic is set to true.
	defer func() {
		if !config.Debug.Panic {
			if err := recover(); err != nil {
				log.Printf("CLIENT PANICKED: %s", c.remoteAddr)
				log.Printf("%s: %s", err, debug.Stack())
			}
		}
		c.close()
	}()

	// Main input processing loop, terminates on any error raised not just read
	// or Parse errors.
	{
		// Variables for use in the loop only hence the scoping outer braces
		var (
			s   = bufio.NewReaderSize(c, inputBuffer) // Sized network read buffer
			err error                                 // function local errors
			in  []byte                                // Input string from buffer
		)

		for c.Error() == nil {
			c.SetReadDeadline(time.Now().Add(config.Server.IdleTimeout))
			if in, err = s.ReadSlice('\n'); err != nil {
				frontend.Zero(in)

				if err != bufio.ErrBufferFull {
					c.SetError(err)
					continue
				}

				for err == bufio.ErrBufferFull {
					in, err = s.ReadSlice('\n')
					frontend.Zero(in)
				}
				c.Write([]byte(text.Bad + "\nYou type too much.\n" + text.Prompt + ">"))
				continue
			}

			fixDEL(&in)
			if err = c.frontend.Parse(in); err != nil {
				c.SetError(err)
			}
			frontend.Zero(in)
		}
	}
}

// fixDEL is used to delete characters when the input contains literal DEL
// characters (ASCII 0x7f or "\b"). This is the case when using a client that
// does not support line editing, for example a plain Windows TELNET client.
//
// For example if you type "ABD" then delete the "D" and enter "C" the data
// sent to the server would be "ABD\bC" if there is no line editing support.
// With line editing "ABC" would be sent to the server.
//
// Calling fixDEL on the data will interpret the DEL characters so that, for
// example, "ABD\bC" becomes "ABC".
//
// It should be noted that this function modifies the slice passed to it.
func fixDEL(in *[]byte) {

	// Start and end offsets for slicing. If a DEL is detected before any text s
	// will be zero, otherwise one. If DEL is detected at the end of the slice
	// with no text following it e will be zero, otherwise one.
	s, e := 0, 0

	for i := bytes.IndexByte(*in, '\b'); i != -1; {
		if i > 0 {
			s = 1
		}
		if i < len(*in) {
			e = 1
		}
		*in = append((*in)[:i-s], (*in)[i+e:]...)

		// Zero out data sliced off so as to not leave recoverable garbage at the
		// end of the slice
		copy((*in)[len(*in):len(*in)+e+s], []byte("\x00\x00"))

		s, e = 0, 0
		i = bytes.IndexByte(*in, '\b')
	}
	return
}

// close shuts down a client cleanly, closes network connections and
// deallocates resources.
func (c *client) close() {

	idle, busy := false, false

	// Idle timeout?
	if oe, ok := c.Error().(*net.OpError); ok && oe.Timeout() {
		idle = true
	}

	// Server busy?
	if _, ok := c.Error().(noLeaseError); ok {
		busy = true
	}

	// Deallocate current frontend if we have one
	if c.frontend != nil {
		if idle {
			c.Write([]byte("\n")) // Move off prompt line
		}
		c.frontend.Close()
		c.frontend = nil
	}

	// If connection timed out notify the client
	if idle {
		c.Write([]byte(text.Bad + "\nIdle connection terminated by server.\n"))
	}

	// Notify if server too busy to accept more players
	if busy {
		c.Write([]byte(text.Bad + "\nServer too busy. Please come back in a short while.\n"))
	}

	// Say goodbye to client
	c.Write([]byte(text.Info + "\nBye bye...\n\n"))

	// Revert to default colors
	c.Write([]byte(text.Reset))

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

	c.leaseRelease()

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
