// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
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
	"code.wolfmud.org/WolfMUD.git/utils/sender"
	"code.wolfmud.org/WolfMUD.git/utils/text"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
	"unicode"
)

// TODO: When we have sorted out global settings some of these need moving
// there.
const (
	MAX_TIMEOUT = 10 * time.Minute // Idle connection timeout
	TERM_WIDTH  = 80               // fold wrapping length - see fold function
	BUFFER_SIZE = 256              // Comms input buffer length in bytes

	GREETING = `

[GREEN]Wolf[WHITE]MUD Copyright 2012 Andrew 'Diddymus' Rolfe

    [GREEN]W[WHITE]orld
    [GREEN]O[WHITE]f
    [GREEN]L[WHITE]iving
    [GREEN]F[WHITE]antasy
`
)

// Client represents a TELNET client connection to the server.
type Client struct {
	parser parser.Interface // Currently attached parser
	name   string           // Current name allocated by attached parser
	conn   *net.TCPConn     // The TELNET network connection
	bail   chan error
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
		conn: conn,
		bail: make(chan error, 1),
	}

	// Initialise bail channel with a nil error
	c.bail <- nil

	c.conn.SetKeepAlive(true)
	c.conn.SetLinger(0)
	c.conn.SetNoDelay(false)

	c.prompt = sender.PROMPT_NONE
	c.Send(GREETING)
	c.prompt = sender.PROMPT_DEFAULT

	c.parser = player.New(c)
	c.name = c.parser.Name()

	log.Printf("Client created: %s", c.name)

	c.receiver()

	c.parser.Destroy()
	c.parser = nil

	if err := c.bailed(); err != nil {
		log.Printf("Comms error for: %s, %s", c.name, err)
	}

	if err := c.conn.Close(); err != nil {
		log.Printf("Error closing socket for %s, %s", c.name, err)
	}

	log.Printf("Spawn ending for %s", c.name)
}

// isBailing checks to see if the client is currently bailing.
func (c *Client) isBailing() bool {
	bailing := <-c.bail
	c.bail <- bailing
	return bailing != nil
}

// bailing records the fact there has been an error and we want to bail. If we
// are already bailing the current error is not overwritten so we always get the
// error that initially caused the client to bail.
func (c *Client) bailing(err error) {
	bailing := <-c.bail
	if bailing == nil {
		bailing = err
	}
	c.bail <- bailing
}

// bailed returns the error that caused the client to be bailing or nil if the
// client is not currently bailing.
func (c *Client) bailed() error {
	bailing := <-c.bail
	c.bail <- bailing
	return bailing
}

// receiver receives data from the user's TELNET client. receive waits on a
// connection for MAX_TIMEOUT before timing out. If the read times out
// the connection will be closed and the inactive user disconnected.
func (c *Client) receiver() {

	// buffer is the input buffer which may be drip fed data from a client. This
	// caters for input being read in multiple reads, multiple inputs being read
	// in a single read and byte-at-a-time reads - currently from Windows telnet
	// clients.  It has a fixed length and capacity to avoid re-allocations.
	// bCursor points to the *next* byte in the buffer to be filled.
	buffer := make([]byte, BUFFER_SIZE, BUFFER_SIZE)
	bCursor := 0

	// Short & simple function to simplify for loop
	//
	// NOTE: Slicing the buffer with buffer[0:bCursor] stops us accidently
	// reading an LF in the garbage portion of the buffer after the cursor.
	// See next TODO for notes on the garbage.
	nextLF := func() int {
		return bytes.IndexByte(buffer[0:bCursor], 0x0A)
	}

	var b int      // bytes read from network
	var err error  // any comms error
	var LF int     // next linefeed position
	var cmd []byte // extracted command to be processed

	// Loop on connection until we bail out or timeout
	for !c.isBailing() && !c.parser.IsQuitting() {

		c.conn.SetReadDeadline(time.Now().Add(MAX_TIMEOUT))
		b, err = c.conn.Read(buffer[bCursor:])

		if b > 0 {

			// If buffer would overflow discard current buffer by
			// setting the buffer length back to zero
			if bCursor+b >= BUFFER_SIZE {
				bCursor = 0
				continue
			}
			bCursor += b

			for LF = nextLF(); LF != -1; LF = nextLF() {

				// NOTE: This could be buffer[0:LF-1] to save trimming the CR before
				// the LF as Telnet is supposed to send CR+LF. However if we just get
				// sent an LF - might be malicious or a badly written / configured
				// client - then [0:LF-1] causes a 'slice bounds out of range' panic.
				// Trimming extra characters is simpler than adding checking
				// specifically for the corner case.

				cmd = bytes.TrimRightFunc(buffer[0:LF], unicode.IsSpace)

				if len(cmd) == 0 {
					c.Send("")
				} else {
					c.parser.Parse(string(cmd))
				}

				// Remove the part of the buffer we just processed by copying the bytes
				// after the bCursor to the front of the buffer.
				//
				// TODO: This has the side effect of being quick and simple but leaves
				// input garbage from the bCursor to the end of the buffer. Will this
				// be an issue? A security issue? We could setup a zero buffer and copy
				// enough to overwrite the garbage? See previous NOTE on garbage check.
				copy(buffer, buffer[LF+1:])
				bCursor -= LF + 1
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
	if !c.isBailing() && !c.parser.IsQuitting() {
		c.prompt = sender.PROMPT_NONE
		c.Send(" ")
		c.parser.Parse("QUIT")
		c.Send("[RED]Idle connection terminated by server.[WHITE]\n")
		log.Printf("Closing idle connection for: %s", c.name)
	}

	log.Printf("receiver ending for %s", c.name)
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

	c.conn.SetWriteDeadline(time.Now().Add(1 * time.Minute))

	dat := []byte(data)
	for len(dat) > 0 {
		if w, err := c.conn.Write(dat); err != nil {
			c.bailing(err)
			return
		} else {
			dat = dat[w:]
		}
	}

	return
}
