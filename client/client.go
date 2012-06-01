// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package client implements a client connecting to the WolfMUD server. It is
// actually a mini TELNET server - any TELNET client should be able to connect
// to and talk to client. It supports ANSI foreground colour codes and wrapping
// on whitespace.
//
// If you take the client package, add some code to accept a connection and pass
// it to a new client instance you practically have a complete but simple TELNET
// server :)
//
// The idea here is to have a client that can talk to any parser. The parser can
// be anything from a login, a way for creating new players, a mini chat system
// or an actual player. A typical example usage might be connect and attach to
// a login parser, once you get a successful login detach the login parser and
// connect a player parser.
//
// You could also detach from your player and attach to mobiles and 'puppet'
// them leading to some interesting possibilities ;)
package client

// BUG(Diddymus): Currently the client package expects TELNET to be in line
// mode - won't work with windows TELNET currently.

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"runtime"
	"strings"
	"time"
	"wolfmud.org/entities/mobile/player"
	"wolfmud.org/utils/broadcaster"
	"wolfmud.org/utils/parser"
)

// TODO: When we have sorted out global settings Interval needs moving there.
const (
	MAX_RETRIES      = 60   // Each retry is 10 seconds
	SEND_BUFFER_SIZE = 4096 // Number of sending messages to buffer
	TERM_WIDTH       = 80   // fold wrapping length - see fold function

	GREETING = `

[GREEN]Wolf[WHITE]MUD © 2012 Andrew 'Diddymus' Rolfe

    [GREEN]W[WHITE]orld
    [GREEN]O[WHITE]f
    [GREEN]L[WHITE]iving
    [GREEN]F[WHITE]antasy


`
)

// colourTable maps colour names to ANSI escape sequences. The sequences are
// defined in the ECMA-48 standard or ISO/IEC 6429.
var colourTable = map[string]string{
	"[BLACK]":   "\033[30m",
	"[RED]":     "\033[31m",
	"[GREEN]":   "\033[32m",
	"[YELLOW]":  "\033[33m", // Note ESC [ 33m can be brown or yellow
	"[BROWN]":   "\033[33m", // So here we have the same code twice
	"[BLUE]":    "\033[34m",
	"[MAGENTA]": "\033[35m",
	"[CYAN]":    "\033[36m",
	"[WHITE]":   "\033[37m",
}

// regexpLF is a package instance Compiled regex to change LF to CR+LF
var regexpLF, _ = regexp.Compile("([^\r])\n")

// Interface should be implemented by anything that want's to be a client and
// interface a TELNET client to the WolfMUD server.
type Interface interface {
	Send(format string, any ...interface{})
	SendWithoutPrompt(format string, any ...interface{})
}

// Client is the default client implementation.
//
// The send channel acts as a demultiplexer serialising and queuing responses
// back to the TELNET client comming from multiple Goroutines.
//
// The senderWakeup channel is used by the receiver Goroutine to wakeup - or
// timeout - the send channel. The receiver times out reading from the network
// connection automatically. If the receiver detects we are bailing it wakes up
// the sender so it too can bail.
type Client struct {
	parser       parser.Interface // Currently attached parser
	name         string           // Current name allocated by attached parser
	conn         *net.TCPConn     // The TELNET network connection
	bail         bool             // Should the client bail and exit?
	send         chan string      // channel queues responses from goroutines
	senderWakeup chan bool        // sender wake up signal
	ending       chan bool        // Used to wait for sender & receiver to end
}

// final is used for debugging to make sure the GC is cleaning up
func final(c *Client) {
	log.Printf("+++ Client %s finalized +++\n", c.name)
}

// Spawn manages the main client Goroutine. It creates the client, starts the
// receiver and sender, waits for them to finish and then cleans up. So it's not
// called New because it does more than create the client. It not called Run or
// Start because it does more than that. Spawn seemed like a good name as it
// spawns a new client and Goroutines :)
//
// TODO: Move display of greeting to login parser.
// TODO: Modify to handle attaching/detatching multiple parsers
func Spawn(conn *net.TCPConn, world broadcaster.Interface) {

	c := &Client{
		conn:         conn,
		send:         make(chan string, SEND_BUFFER_SIZE),
		senderWakeup: make(chan bool, 1),
		ending:       make(chan bool),
	}

	c.SendWithoutPrompt(GREETING)

	c.parser = player.New(c, world)
	c.name = c.parser.Name()

	log.Printf("Client created: %s\n", c.name)
	runtime.SetFinalizer(c, final)

	go c.receiver()
	go c.sender()

	<-c.ending
	<-c.ending

	c.parser.Destroy()
	c.parser = nil

	if err := c.conn.Close(); err != nil {
		log.Printf("Error closing socket for %s, %s\n", c.name, err)
	}

	log.Printf("Spawn ending for %s\n", c.name)
}

// receiver is run as a Goroutine to receive data from the user's TELNET
// client. receive waits on a connection for 10 seconds before timing out.
// At this point it decrements the idleRetrys counter. If idleRetrys reaches
// zero the connection will be closed and the inactive user disconnected. Any
// received data resets the idleRetrys to the value of MAX_RETRIES. This means
// that and idle session will be disconnected after MAX_RETRIES * 10 seconds.
//
// The other half of the connection is handled by the sender Goroutine.
func (c *Client) receiver() {

	var inBuffer [255]byte

	c.conn.SetKeepAlive(false)
	c.conn.SetLinger(0)
	idleRetrys := MAX_RETRIES

	// Loop on connection until we bail out or run out of retries
	for ; !c.bail && idleRetrys > 0; idleRetrys-- {
		c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		if b, err := c.conn.Read(inBuffer[0:254]); err != nil {
			if oe, ok := err.(*net.OpError); !ok || !oe.Timeout() {
				log.Printf("Comms error for: %s, %s\n", c.name, err)
				c.bail = true
			}
		} else {
			input := strings.TrimSpace(string(inBuffer[0:b]))
			c.parser.Parse(input)
			if c.parser.Quitting() {
				c.bail = true
			}
			idleRetrys = MAX_RETRIES + 1
		}
	}

	// Connection idle and we ran out of retries?
	if idleRetrys == 0 {
		c.SendWithoutPrompt("\n\n[RED]Idle connection terminated by server.\n\n[YELLOW]Bye Bye[WHITE]\n\n")
		log.Printf("Closing idle connection for: %s\n", c.name)
		c.bail = true
	}

	log.Printf("Sending wakeup signal for %s\n", c.name)
	c.senderWakeup <- true

	log.Printf("receiver ending for %s\n", c.name)
	c.ending <- true
}

func (c *Client) Send(format string, any ...interface{}) {
	c.SendWithoutPrompt("[WHITE]"+format+"\n[MAGENTA]>", any...)
}

func (c *Client) SendWithoutPrompt(format string, any ...interface{}) {
	if c.bail {
		//log.Printf("oops %s dropping message %s\n", c.name, fmt.Sprintf(format, any...))
	} else {
		if (cap(c.send) - len(c.send)) < 5 {
			log.Printf("oops %s dropping message, sending too slow.\n", c.name)
		} else {

			// NOTE: You need to colourize THEN fold so fold counts colour codes
			// and NOT colour names ;)
			data := fmt.Sprintf(format, any...)
			data = colourize(data)
			data = fold(data)
			data = regexpLF.ReplaceAllString(data, "$1\r\n")

			c.send <- data
		}
	}
}

// sender is run as a Goroutine to send data to the user's TELNET client. Unlike
// the receiver Goroutine this method blocks reading from the send channel which
// is used to serialise multiple messages arriving from multiple Goroutines.
// Due to this a boolean can be sent on the senderWakeup channel to 'timeout'
// the blocking. This is commonly used when the receiver Goroutine notices a
// problem, sets the bail flag and then wakes up the sender so that it takes
// notice and ends.
func (c *Client) sender() {

	for !c.bail {
		select {
		case <-c.senderWakeup:
			c.bail = true
			break
		case msg := <-c.send:
			if c.bail {
				//log.Printf("oops %s dropping message %s\n", c.name, msg)
			} else {
				if _, err := c.conn.Write([]byte(msg)); err != nil {
					log.Printf("Comms error for: %s, %s\n", c.name, err)
					c.bail = true
					break
				}
			}
		}
	}

	log.Printf("sender ending for %s\n", c.name)
	c.ending <- true
}

// fold takes a string of text and turns its into lines of TERM_WIDTH length
// breaking on whitespace. The text may contain ANSI colour codes in the format
// \033[xxm - for values of xx see the definition of colourTable. Line endings
// are expected to be Linefeeds only - LF, \n or 0x0A - normal for *nix systems.
//
// TODO: Softcode TERM_WIDTH via a user/player setting.
//
// TODO: Assumes control sequences are 5 bytes.
//
// TODO: Could probably use some Unicode love.
//
// TODO: Needs to be optimized.
func fold(in string) (out string) {
	p := 0
	for _, word := range strings.SplitAfter(in, " ") {
		for _, atom := range strings.SplitAfter(word, "\n") {
			l := len(atom) - strings.Count(atom, "\n") - (strings.Count(atom, "\033") * 5)
			if p+l > TERM_WIDTH {
				out += "\n"
				p = 0
			}
			p = p + l
			if strings.HasSuffix(atom, "\n") {
				p = 0
			}
			out += atom
		}
	}
	return
}

// colourize turns colour names into colour ANSI codes within a string. This
// allows messages to be coloured easily with colour names. For example the
// message:
//
//	"[RED]Boom![WHITE]"
//
// will be turned into:
//
//	"\033[31mBoom!\033[37m"
//
// Ultimately printing "Boom!" in red. Messages do not need to end in "[WHITE]"
// as this will be added automatically so you can't forget to do it. Colours
// can be changed as many times as you want:
//
//	"[RED]C[GREEN]o[YELLOW]l[BLUE]o[MAGENTA]u[CYAN]r"
//
// Prints "Colours" each letter in a different colour.
//
// TODO: Extend to include background colours?
func colourize(in string) (out string) {
	for colour, code := range colourTable {
		in = strings.Replace(in, colour, code, -1)
	}
	return in
}

// monochrome strips colour names from a string. This function is like
// colourize except the colour name is replaced with nothing - in effect
// stripping the colours.
func monochrome(in string) (out string) {
	for colour := range colourTable {
		in = strings.Replace(in, colour, "", -1)
	}
	return in
}
