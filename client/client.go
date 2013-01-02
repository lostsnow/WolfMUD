// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package client implements a client connecting to the WolfMUD server. It is
// actually a mini TELNET server - any TELNET client should be able to connect
// to and talk to client. It supports ANSI foreground color codes and wrapping
// on whitespace.
//
// If you take the client package, write some code to accept a connection and
// pass it to client.Spawn you practically have a simple TELNET server that can
// be extended and used for a number of projects :)
//
// The idea here is to have a client that can talk to any parser. The parser
// can be anything from a login, a menu system, a mini chat system or an actual
// player session. A typical example usage might be connect and attach to a
// login parser, once you get a successful login detach the login parser and
// connect a player parser.
//
// You could also detach from your player's parser and attach to a mobile's
// parser and 'puppet' them leading to some interesting possibilities ;)
package client

// BUG(Diddymus): Currently we don't try to put the client into line mode. If
// the client is in character at a time mode it will work but be inefficient.
// This should be sorted out when RFC1184 is implemented along with a *lot* of
// other TELNET related RFCs.

import (
	"bytes"
	"code.wolfmud.org/WolfMUD.git/entities/mobile/player"
	"code.wolfmud.org/WolfMUD.git/utils/parser"
	"code.wolfmud.org/WolfMUD.git/utils/text"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// TODO: When we have sorted out global settings some of these need moving
// there.
const (
	MAX_TIMEOUT = 10 * time.Minute // Idle connection timeout
	TERM_WIDTH  = 80               // fold wrapping length - see fold function

	GREETING = `

[GREEN]Wolf[WHITE]MUD Copyright 2012 Andrew 'Diddymus' Rolfe

    [GREEN]W[WHITE]orld
    [GREEN]O[WHITE]f
    [GREEN]L[WHITE]iving
    [GREEN]F[WHITE]antasy
`
)

// Prompt definitions
const (
	PROMPT_NONE    = "\n"
	PROMPT_DEFAULT = text.COLOR_MAGENTA + "\n>"
)

// Client represents a TELNET client connection to the server.
type Client struct {
	parser parser.Interface // Currently attached parser
	name   string           // Current name allocated by attached parser
	conn   *net.TCPConn     // The TELNET network connection
	bail   error
	mutex  chan bool
	prompt string
}

// Spawn manages the client Goroutine which is normally launched by the World.
// It creates the client, starts the receiver, waits for it to finish and then
// cleans up. So it's not called New because it does more than create the
// client. It not called Run or Start because it does more than that. Spawn
// seemed like a good name as it spawns a new client :)
//
// TODO: Move display of greeting to login parser.
//
// TODO: Modify to handle attaching/detatching multiple parsers
func Spawn(conn *net.TCPConn) {

	c := &Client{
		conn:  conn,
		mutex: make(chan bool, 1),
	}

	c.prompt = PROMPT_NONE
	c.Send(GREETING)
	c.prompt = PROMPT_DEFAULT

	c.parser = player.New(c)
	c.name = c.parser.Name()

	log.Printf("Client created: %s\n", c.name)

	c.receiver()

	c.parser.Destroy()
	c.parser = nil

	if c.bail != nil {
		log.Printf("Comms error for: %s, %s\n", c.name, c.bail)
	}

	if err := c.conn.Close(); err != nil {
		log.Printf("Error closing socket for %s, %s\n", c.name, err)
	}

	log.Printf("Spawn ending for %s\n", c.name)
}

// Lock takes the lock on the client
func (c *Client) lock() {
	c.mutex <- true
}

// Unlock releases the lock on the client
func (c *Client) unlock() {
	<-c.mutex
}

// isBailing checks to see if the client is currently bailing.
func (c *Client) isBailing() bool {
	c.lock()
	defer c.unlock()
	return c.bail != nil
}

// bailing records the fact there has been an error and we want to bail. If we
// are already bailing the current error is not overwritten so we always get the
// error that initially caused the client to bail.
func (c *Client) bailing(err error) {
	c.lock()
	defer c.unlock()
	if c.bail == nil {
		c.bail = err
	}
}

// receiver receives data from the user's TELNET client. receive waits on a
// connection for MAX_TIMEOUT before timing out. If the read times out
// the connection will be closed and the inactive user disconnected.
func (c *Client) receiver() {

	const inBuffLen = 32

	var inBuffer [inBuffLen]byte

	lineBuffer := []byte{}

	// Short & simple function to simplify for loop
	nextLF := func() int {
		return bytes.IndexByte(lineBuffer, 0x0A)
	}

	c.conn.SetKeepAlive(true)
	c.conn.SetLinger(0)
	c.conn.SetNoDelay(false)

	var b int      // bytes read from network
	var err error  // any comms error
	var LF int     // next linefeed position
	var cmd []byte // extracted command to be processed

	// Loop on connection until we bail out or timeout
	for !c.isBailing() && !c.parser.IsQuitting() {

		c.conn.SetReadDeadline(time.Now().Add(MAX_TIMEOUT))
		b, err = c.conn.Read(inBuffer[0 : inBuffLen-1])

		if b > 0 {
			lineBuffer = append(lineBuffer, inBuffer[0:b]...)

			for LF = nextLF(); LF != -1; LF = nextLF() {

				// NOTE: This could be lineBuffer[0:LF-1] to save TrimSpaceing the CR
				// before the LF as Telnet is supposed to send CR+LF. However if we
				// just get sent an LF - might be malicious or a badly written /
				// configured client - then [0:LF-1] causes a 'slice bounds out of
				// range' panic. Trimming an extra character is simpler than adding
				// checking specifically for the corner case.

				cmd = bytes.TrimSpace(lineBuffer[0:LF])

				if len(cmd) == 0 {
					c.Send("")
				} else {
					c.parser.Parse(string(cmd))
				}

				// Remove the part of the buffer we just processed
				lineBuffer = lineBuffer[LF+1:]
			}
		}

		// Check for errors reading data (see io.Reader for details)
		if err != nil {
			if oe, ok := err.(*net.OpError); !ok || !oe.Timeout() {
				c.bailing(err)
			}
			break
		}

	}

	// If we are not quitting or bailing we timed out
	if !c.parser.IsQuitting() && !c.isBailing() {
		c.prompt = PROMPT_NONE
		c.Send(" ")
		c.parser.Parse("QUIT")
		c.Send("[RED]Idle connection terminated by server.[WHITE]\n")
		log.Printf("Closing idle connection for: %s\n", c.name)
	}

	log.Printf("receiver ending for %s\n", c.name)
}

// Prompt sets a new prompt and returns the old prompt. It is implemented as
// part of the Sender interface.
func (c *Client) Prompt(newPrompt string) (oldPrompt string) {
	oldPrompt, c.prompt = c.prompt, newPrompt
	return
}

// Send takes a message with parameters, adds a prompt and sends the message on
// it's way to the client. Send is modelled after the fmt.Sprintf function and
// takes a format string and parameters in the same way. In addition the current
// prompt is added to the end of the message.
//
// If the format string is empty we can take a shortcut and just redisplay the
// prompt. Otherwise we process the whole enchilada making sure the prompt is on
// a new line when displayed.
//
// NOTE: Send can be called by multiple goroutines.
func (c *Client) Send(format string, any ...interface{}) {
	if c.isBailing() {
		return
	}

	data := text.COLOR_WHITE + format + c.prompt

	if len(any) > 0 {
		data = fmt.Sprintf(data, any...)
	}

	// NOTE: You need to colorize THEN fold so fold counts the length of color
	// codes and NOT color names ;)
	data = text.Fold(text.Colorize(data), TERM_WIDTH)
	data = strings.Replace(data, "\n", "\r\n", -1)

	if _, err := c.conn.Write([]byte(data)); err != nil {
		c.bailing(err)
	}

	return
}
