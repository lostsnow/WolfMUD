// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/text"

	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

// TODO: These need to be configuration options once we have them
const (
	timeout       = 10 * time.Minute // Idle connection timeout
	terminalWidth = 80               // fold wrapping length - see fold function
	lineBuffer    = 80               // Default network read & line buffer size
	pageBuffer    = 80 * 24          // Default network write buffer size
)

// Values to be treated as constants but we can't define them as constants
var (
	crlf          = []byte("\r\n") // Carriage return/line feed for network data
	lf            = crlf[1:]       // Line feed used internally for data
	defaultPrompt = []byte("\n>")  // Default prompt for client input

	// Most of the flow and control for the client is done using errors so we
	// raise an "I want to quit" error instead of adding another level of
	// checking for a separate quitting flag
	quitting = errors.New("Client quitting error")
)

// client contains state information about a client connection. The err field
// should not be manipulated directly. Instead call Error() and SetError().
//
// TODO: client is currently talking directly to a player. It should be talking
// to a switchable, abstract layer so that we can talk to a player, menus,
// account system etc.
type client struct {
	*net.TCPConn            // The client's connection
	err          chan error // Error channel to sync between input & output
	player       has.Thing  // The player this client is associated with
	prompt       []byte     // The current prompt the client is using
}

// nextPlayerID is used to get the next available unique player ID
var nextPlayerID <-chan []byte

// Temporary unique player ID generator used to create "Player x" names and
// "PLAYERx" aliases until we have accounts up and running
func init() {
	c := make(chan []byte)
	nextPlayerID = c
	go func() {
		playerID := []byte("0000001")
		for {
			c <- playerID
			for p := 6; p >= 0; p-- {
				playerID[p]++
				if playerID[p] <= '9' {
					break
				}
				playerID[p] = '0'
			}
		}
	}()
}

// newClient returns an initialised client for the passed connection.
func newClient(conn *net.TCPConn) *client {

	// Setup connection parameters
	conn.SetKeepAlive(true)
	conn.SetLinger(0)
	conn.SetNoDelay(false)
	conn.SetWriteBuffer(pageBuffer)
	conn.SetReadBuffer(lineBuffer)

	id := string(<-nextPlayerID)

	c := &client{
		TCPConn: conn,
		err:     make(chan error, 1),
		prompt:  defaultPrompt,

		// Setup test player
		player: attr.NewThing(
			attr.NewName("Player "+id),
			attr.NewAlias("PLAYER"+id),
			attr.NewInventory(),
			attr.NewLocate(nil),
		),
	}

	c.err <- nil

	// Add player attribute with reference to client for sending back data
	c.player.Add(attr.NewPlayer(c))

	// Put player into the world
	if i := attr.FindInventory(world["loc1"]); i != nil {
		i.Lock()
		i.Add(c.player)
		i.Unlock()
	}

	// Describe what they can see
	msg, _ := cmd.Parse(c.player, "LOOK")
	c.Write(msg)

	return c
}

// process handles input from the network connection.
func (c *client) process() {

	var (
		s   = bufio.NewReaderSize(c, lineBuffer) // Sized network reading buffer
		err error                                // function local errors
		in  string                               // Input string from buffer
	)

	// Main input processing loop
	for err == nil {
		c.SetReadDeadline(time.Now().Add(timeout))
		in, err = s.ReadString('\n')

		// Do we need to set an error?
		if err != nil {
			c.SetError(err)
			break
		}

		// Anyone else set an error?
		if c.Error() != nil {
			break
		}

		// Process the input, if we get an error the loop will exit
		if msg, _ := cmd.Parse(c.player, in); len(msg) > 0 {
			c.Write(msg)
		}

		// Remember ReadString will return the delimiters...
		if strings.TrimSpace(strings.ToUpper(in)) == "QUIT" {
			c.SetError(quitting)
			break
		}

	}

	// Log reson for ending, notify player if we can
	//
	// NOTE: We do not log EOF with no input otherwise the log can get very
	// noisy. We also report EOF seperatly so we can log the host and socket.
	//
	// TODO: Log can still get noisy with errors. Might add a configure knob to
	// just log quits, timeouts and drops which is what you usually want to know
	// if trying to handle a player dispute ;)
	switch err := c.Error(); {
	case err == quitting:
		log.Printf("Quit received from: %s", c.RemoteAddr())
	case err == io.EOF:
		if in != "" {
			log.Printf("Connection error: %s %s", c.RemoteAddr(), err)
		}
	case err != nil:
		if oe, ok := err.(*net.OpError); ok && oe.Timeout() {
			log.Printf("Connection timeout: %s", c.RemoteAddr())
			c.Write([]byte("quit\n\nIdle connection terminated by server."))
		} else {
			log.Printf("Connection error: %s", err)
		}
	default:
		log.Printf("Connection dropped by: %s", c.RemoteAddr())
	}

	// If not voluntarily quitting do it automatically
	if c.Error() != quitting {
		msg, _ := cmd.Parse(c.player, "QUIT")
		c.Write(msg)
	}

	// Make sure connection closed down
	if err = c.Close(); err != nil {
		log.Printf("Error closing connection: %s", err)
	} else {
		log.Printf("Connection closed: %s", c.RemoteAddr())
	}
	c.TCPConn = nil

	// Remove cyclic reference
	if a := attr.FindPlayer(c.player); a != nil {
		c.player.Remove(a)
	}

	// Close and drain channel
	close(c.err)
	<-c.err

	return
}

// Write handles output for the network connection.
func (c *client) Write(d []byte) (n int, err error) {

	// Don't try doing anything if we already have errors
	if c.Error() != nil {
		return
	}

	d = append(d, c.prompt...)
	t := text.Fold(d, terminalWidth)

	c.SetWriteDeadline(time.Now().Add(timeout))

	if n, err = c.TCPConn.Write(bytes.Replace(t, lf, crlf, -1)); err != nil {
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
