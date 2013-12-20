// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package client implements asynchronous network I/O for a client connecting
// to the WolfMUD server. It will handle network errors and idle timeouts
// gracefully.
package client

// BUG(Diddymus): Currently we don't try to put the client into line mode. If
// the client is in character at a time mode it will work but be inefficient.
// This should be sorted out when RFC1184 is implemented along with a *lot* of
// other TELNET related RFCs.

import (
	"code.wolfmud.org/WolfMUD.git/driver"
	"code.wolfmud.org/WolfMUD.git/utils/sender"
	"code.wolfmud.org/WolfMUD.git/utils/text"

	"bytes"
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
)

// Client represents a client connection to the server. We could embed the
// *net.TCPConn but do we really want to expose all of it's functionality via a
// Client?
type Client struct {
	conn   *net.TCPConn
	bail   chan error
	prompt string
}

// Spawn manages the client Goroutine which is normally launched by the World.
// It creates the client, starts the receiver, waits for it to finish and then
// cleans up. So it's not called New because it does more than create the
// client. It not called Run or Start because it does more than that. Spawn
// seemed like a good name as it spawns a new client :)
func Spawn(conn *net.TCPConn) {

	c := &Client{
		conn:   conn,
		bail:   make(chan error, 1),
		prompt: sender.PROMPT_DEFAULT,
	}

	// Initialise bail channel with a nil error
	c.bail <- nil

	c.conn.SetKeepAlive(true)
	c.conn.SetLinger(0)
	c.conn.SetNoDelay(false)

	c.receiver()

	if err := c.bailed(); err != nil {
		log.Printf("Comms error for: %s, %s", c, err)
	} else {
		c.Prompt(sender.PROMPT_NONE)
		c.Send("[YELLOW]Bye Bye[WHITE]\n")
	}

	if err := c.conn.Close(); err != nil {
		log.Printf("Error closing socket for: %s, %s", c, err)
	} else {
		log.Printf("Clean player exit, socket closed for: %s", c)
	}
}

// String returns a client identifier. Currently this has the format of:
//
//	remote_address:remote_port
//
// Having it as a function here makes it easy to change later on if we want to.
func (c *Client) String() string {
	return c.conn.RemoteAddr().String()
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

// receiver loops reading data from the client's network connection and writing
// it to the current driver for processing. If the read times out the connection
// will be closed and the inactive user disconnected.
func (c *Client) receiver() {

	// Our initial login driver.
	driver := driver.New(c)

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
	for !c.isBailing() && !driver.IsQuitting() {

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

				driver.Process(string(cmd))

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
			if oe, ok := err.(*net.OpError); ok && oe.Timeout() {
				c.prompt = sender.PROMPT_NONE
				c.Send("")
				driver.Logout()
				c.Send("\n\n[RED]Idle connection terminated by server.")
				log.Printf("Closing idle connection for: %s", c)
				break
			}
			c.bailing(err)
		}

	}

	driver.Logout()
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
