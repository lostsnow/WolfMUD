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

// BUG(Diddymus): Currently the client package expects TELNET to be in line
// mode - won't work with windows TELNET currently.
// UPDATE: Will work if KLUDGE set to true which enables a nasty kludge :(

import (
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
	MAX_TIMEOUT = 10 // Idle connection timeout in minutes
	TERM_WIDTH  = 80 // fold wrapping length - see fold function

	// Set to true if using Windows telnet or some other nasty client that
	// defaults to character-at-a-time mode and not linemode.
	KLUDGE = false

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
	PROMPT_NONE    = ""
	PROMPT_DEFAULT = text.COLOR_MAGENTA + ">"
)

// Client is the default client implementation.
//
// The Client type implements the sender interface.
//
// The send channel acts as a demultiplexer serialising and queuing responses
// back to the TELNET client comming from multiple Goroutines.
//
// The senderWakeup channel is used by the receiver Goroutine to wakeup - or
// timeout - the send channel. The receiver times out reading from the network
// connection automatically. If the receiver detects we are bailing it wakes up
// the sender so it too can bail.
type Client struct {
	parser parser.Interface // Currently attached parser
	name   string           // Current name allocated by attached parser
	conn   *net.TCPConn     // The TELNET network connection
	bail   error
	mutex  chan bool
	prompt string
}

// Spawn manages the main client Goroutine. It creates the client, starts the
// receiver, waits for it to finish and then cleans up. So it's not called New
// because it does more than create the client. It not called Run or Start
// because it does more than that. Spawn seemed like a good name as it spawns a
// new client and Goroutine :)
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

// receiver is run as a Goroutine to receive data from the user's TELNET client.
// receive waits on a connection for MAX_TIMEOUT minutes before timing out.
// If the read times out the connection will be closed and the inactive user
// disconnected.
func (c *Client) receiver() {

	var inBuffer [255]byte

	// Only needed if KLUDGE = true :(
	var tempBuff [255]byte
	tempOff := 0

	c.conn.SetKeepAlive(true)
	c.conn.SetLinger(0)
	c.conn.SetNoDelay(false)

	// Loop on connection until we bail out or timeout
	for !c.isBailing() && !c.parser.IsQuitting() {
		c.conn.SetReadDeadline(time.Now().Add(MAX_TIMEOUT * time.Minute))

		if b, err := c.conn.Read(inBuffer[0:254]); err != nil {
			if oe, ok := err.(*net.OpError); !ok || !oe.Timeout() {
				c.bailing(err)
			}
			break
		} else {

			// KLUDGE enables a temporary buffer to be built up until we see an LF at
			// which point we swap our temporary buffer for the partial buffer and
			// pretend nothing weird happened...
			//
			// THIS IS NOT NICE, NOT SAFE, NOT FUNNY, NOT CHEESE! USE AT OWN RISK -
			// MAY DESTROY YOUR COMPUTER, COFFEE MAKER AND THE UNIVERSE - not
			// necessarily in that order or any other order.
			//
			// This should not be needed when RFC1184 is implemented along with a
			// *lot* of other TELNET related RFCs
			if KLUDGE {
				copy(tempBuff[tempOff:tempOff+b], inBuffer[0:b])
				tempOff += b
				if tempBuff[tempOff-1] != 0x0A {
					continue
				} else {
					copy(inBuffer[0:tempOff], tempBuff[0:tempOff])
					b, tempOff = tempOff, 0
				}
			}

			input := strings.TrimSpace(string(inBuffer[0:b]))
			if input == "" {
				c.Send("")
			} else {
				c.parser.Parse(input)
			}
		}
	}

	// If we are not quitting we timed out
	if !c.parser.IsQuitting() {
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

	if len(format) > 0 {
		format += "\n"
	}

	any = append(any, c.prompt)
	data := text.COLOR_WHITE + format + "%s"

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
