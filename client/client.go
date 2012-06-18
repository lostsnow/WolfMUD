// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package client implements a client connecting to the WolfMUD server. It is
// actually a mini TELNET server - any TELNET client should be able to connect
// to and talk to client. It supports ANSI foreground colour codes and wrapping
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
	"fmt"
	"log"
	"net"
	"regexp"
	"runtime"
	"strings"
	"time"
	"wolfmud.org/entities/location/startingLocation"
	"wolfmud.org/entities/mobile/player"
	"wolfmud.org/utils/parser"
)

// TODO: When we have sorted out global settings some of these need moving
// there.
const (
	MAX_RETRIES = 60 // Each retry is 10 seconds
	TERM_WIDTH  = 80 // fold wrapping length - see fold function

	// Set to true if using Windows telnet or some other nasty client that
	// defaults to character-at-a-time mode and not linemode.
	KLUDGE = false

	GREETING = `

[GREEN]Wolf[WHITE]MUD © 2012 Andrew 'Diddymus' Rolfe

    [GREEN]W[WHITE]orld
    [GREEN]O[WHITE]f
    [GREEN]L[WHITE]iving
    [GREEN]F[WHITE]antasy


`
)

// Prompt definitions
const (
	PROMPT_NONE    = ""
	PROMPT_DEFAULT = "[MAGENTA]>"
)

// colourTable maps colour names to ANSI escape sequences. The sequences are
// defined in the ECMA-48 standard or ISO/IEC 6429.
//
// TODO: Add more codes like background colours, underline, bold, normal ???
var colourTable = map[string]string{
	"[BLACK]":   "\033[30m",
	"[RED]":     "\033[31m",
	"[GREEN]":   "\033[32m",
	"[YELLOW]":  "\033[33m", // Note ESC [ 33m can be brown or yellow
	"[BROWN]":   "\033[33m", // So here we have the same escape code twice
	"[BLUE]":    "\033[34m",
	"[MAGENTA]": "\033[35m",
	"[CYAN]":    "\033[36m",
	"[WHITE]":   "\033[37m",
}

// regexpLF is a package instance compiled regex to change LF to CR+LF
var regexpLF, _ = regexp.Compile("([^\r])\n")

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

// final is used for debugging to make sure the GC is cleaning up
func final(c *Client) {
	log.Printf("+++ Client %s finalized +++\n", c.name)
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
func Spawn(conn *net.TCPConn, l *startingLocation.StartingLocation) {

	c := &Client{
		conn:  conn,
		mutex: make(chan bool, 1),
	}

	c.prompt = PROMPT_NONE
	c.Send(GREETING)
	c.prompt = PROMPT_DEFAULT

	c.parser = player.New(c, l)
	c.name = c.parser.Name()

	log.Printf("Client created: %s\n", c.name)
	runtime.SetFinalizer(c, final)

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

// receiver is run as a Goroutine to receive data from the user's TELNET
// client. receive waits on a connection for 10 seconds before timing out.
// At this point it decrements the idleRetrys counter. If idleRetrys reaches
// zero the connection will be closed and the inactive user disconnected. Any
// received data resets the idleRetrys to the value of MAX_RETRIES. This means
// that and idle session will be disconnected after MAX_RETRIES * 10 seconds.
func (c *Client) receiver() {

	var inBuffer [255]byte

	// Only needed if KLUDGE = true :(
	var tempBuff [255]byte
	tempOff := 0

	c.conn.SetKeepAlive(false)
	c.conn.SetLinger(0)
	idleRetrys := MAX_RETRIES

	// Loop on connection until we bail out or run out of retries
	for ; !c.isBailing() && !c.parser.IsQuitting() && idleRetrys > 0; idleRetrys-- {
		c.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		if b, err := c.conn.Read(inBuffer[0:254]); err != nil {
			if oe, ok := err.(*net.OpError); !ok || !oe.Timeout() {
				c.bailing(err)
				break
			}
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
			idleRetrys = MAX_RETRIES + 1
		}
	}

	// Connection idle and we ran out of retries?
	if idleRetrys == 0 {
		c.prompt = PROMPT_NONE
		c.Send("\n\n[RED]Idle connection terminated by server.\n\n[YELLOW]Bye Bye[WHITE]\n\n")
		log.Printf("Closing idle connection for: %s\n", c.name)
	}

	log.Printf("receiver ending for %s\n", c.name)
}

// Send takes a message with parameters, adds a prompt and sends the message on
// it's way to the client. Send is modelled after the fmt.Sprintf function and
// takes a format string and parameters in the same way. In addition the current
// prompt is added to the end of the message.
//
// If the format string is empty we can take a shortcut and just redisplay the
// prompt. Otherwise we process the whole enchilada making sure the prompt is on
// a new line when displayed.
func (c *Client) Send(format string, any ...interface{}) {
	if c.isBailing() {
		return
	}

	if len(format) > 0 {
		format += "\n"
	}

	any = append(any, c.prompt)
	format = "[WHITE]" + format + "%s"

	// NOTE: You need to colourize THEN fold so fold counts the length of colour
	// codes and NOT colour names ;)
	data := fmt.Sprintf(format, any...)
	data = fold(colourize(data))
	data = regexpLF.ReplaceAllString(data, "$1\r\n")

	if _, err := c.conn.Write([]byte(data)); err != nil {
		c.bailing(err)
	}

	return
}

// BUG(Diddymus): fold assumes control sequences are 5 bytes long. When we add
// more control sequences they probably won't be 5 bytes long.

// fold takes a string of text and turns it into lines of TERM_WIDTH length
// breaking on whitespace. The text may contain ANSI colour codes in the format
// \033[xxm - for values of xx see the definition of colourTable. Line endings
// are expected to be Linefeeds only - LF, \n or 0x0A - common on *nix systems.
//
// TODO: Softcode TERM_WIDTH via a user/player setting.
//
// TODO: Could probably use some Unicode love.
//
// TODO: Needs to be optimized.
func fold(in string) (out string) {

	// Shortcut
	if len(in) < TERM_WIDTH {
		return in
	}

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
// Prints "Colour" each letter in a different colour.
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
