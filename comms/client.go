// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms

import (
	"bufio"
	"io"
	"net"
	"runtime/debug"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/frontend"
	"code.wolfmud.org/WolfMUD.git/log"
	"code.wolfmud.org/WolfMUD.git/text"
)

// TODO: These need to be configuration options once we have them
const (
	termColumns  = 80
	termLines    = 24
	inputBuffer  = 512
	writeTimeout = time.Second * 10
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
	err          chan error // Error channel to sync between input & output
	log          log.Conn   // Connection specific logger

	frontend interface { // The current frontend in use
		Parse([]byte) error
		Close()
	}
}

// newClient returns an initialised client for the passed connection.
func newClient(conn *net.TCPConn, seq uint64) *client {

	// Setup connection parameters
	conn.SetKeepAlive(true)
	conn.SetLinger(10)
	conn.SetNoDelay(false)
	conn.SetWriteBuffer(termColumns * termLines)
	conn.SetReadBuffer(inputBuffer)

	c := &client{
		TCPConn: conn,
		err:     make(chan error, 1),
		log:     log.NewConn(seq),
	}

	c.err <- nil
	c.leaseAcquire()

	// Setup frontend if no error acquiring a lease
	if c.Error() == nil {
		c.frontend = frontend.New(c.log, c)
		if config.Server.LogClient {
			c.log("connection from %s", conn.RemoteAddr().String())
		}
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
				c.log("CLIENT PANICKED:")
				c.log("%s: %s", err, debug.Stack())
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

			clean(&in)
			if err = c.frontend.Parse(in); err != nil {
				c.SetError(err)
			}
			frontend.Zero(in)
		}
	}
}

// clean is used to clean up and validate incoming data from clients. The data
// to be cleaned should be passed in a *[]byteslice which will be modified by
// the function call to contain only cleaned data.
//
// For all C0 and C1 control codes, other than BS (backspace, ASCII 0x08,
// '\b'), we drop the byte. This includes any line feeds or carriage returns.
//
//
//               C0 Block         C1 Block
//           ---------------  ---------------
//    ASCII    0x00 - 0x1F      0x80 - 0x9F
//    UTF-8    0x00 - 0x1F    0xC280 - 0xC29F
//  Unicode  U+0000 - U+001F  U+0080 - U+009F
//
//
// BS has special handling, with the control code DEL (delete, ASCII 0x7F)
// being treated the same as BS. If an invalid UTF-8 encoding is found the
// offending bytes will be dropped.
//
// SPECIAL HANDLING FOR BS/DEL
//
// Input from a client may contain literal BS or DEL control codes when using a
// client that does not support line editing, for example a plain Windows
// TELNET client.
//
// For example if you type "ABD" then delete the "D" and enter "C" the data
// sent to the server would be "ABD\bC" if there is no line editing support.
// With line editing "ABC" would be sent to the server.
//
// Calling clean on the data will interpret the BS and DEL control codes so
// that, for example, "ABD\bC" becomes "ABC".
func clean(in *[]byte) {

	const (
		BS    = 0x08     // Backspace
		LoASC = 0x20     // Start of printable ASCII range
		HiASC = 0x7E     // End of printable ASCII range
		DEL   = 0x7F     // Delete
		LoC1  = '\u0080' // Start of C1 control codes
		HiC1  = '\u009F' // End of C1 control codes
	)

	data := *in                    // Dereference input slice
	w := 0                         // Data writing position
	B := [utf8.UTFMax]byte{}       // Rune as bytes temp buffer
	Z := make([]byte, utf8.UTFMax) // Zeroes for clearing data

	for r, ld := 0, len(data); r < ld; { // r = Data read position

		/*
			FAST PATH - for handling simple ASCII bytes
		*/

		// Simple single byte printable ASCII
		if LoASC <= data[r] && data[r] <= HiASC {
			data[r], data[w] = 0x00, data[r]
			w++
			r++
			continue
		}

		// Simple delete of single byte printable ASCII
		if data[r] == BS || data[r] == DEL {
			// If nothing to delete just drop BS/DEL
			if w == 0 {
				data[r] = 0x00
				r++
				continue
			}
			// If deleting single byte ASCII remove it and drop BS/DEL
			if LoASC <= data[w-1] && data[w-1] <= HiASC {
				w--
				data[r], data[w] = 0x00, 0x00
				r++
				continue
			}
		}

		// Drop C0 control codes, except BS and DEL
		if data[r] < LoASC && data[r] != BS && data[r] != DEL {
			data[r] = 0x00
			r++
			continue
		}

		/*
		 SLOW PATH - have to deal with multibyte UTF-8
		*/

		// Get current rune and length, save it's UTF-8 bytes, drop from input
		R, L := utf8.DecodeRune(data[r:])
		copy(B[:L], data[r:r+L])
		copy(data[r:r+L], Z[:L])
		r += L

		switch {
		case R == utf8.RuneError:
			// Drop invalid rune
		case R == BS || R == DEL:
			// Handle delete, repeat until non-combining rune found
			for comb := true; comb; {
				// Decode last rune written in data[:w] taking advantage of the fact we
				// know it's valid already and we just want the length and rune
				switch {
				case data[w-1]&0x80 == 0x00:
					L = 1
					R = rune(data[w-1])
				case data[w-2]&0xC0 == 0xC0:
					L = 2
					R = rune(data[w-2]&0x1F)<<6 |
						rune(data[w-1]&0x3F)
				case data[w-3]&0xC0 == 0xC0:
					L = 3
					R = rune(data[w-3]&0x0F)<<12 |
						rune(data[w-2]&0x3F)<<6 |
						rune(data[w-1]&0x3F)
				case data[w-4]&0xC0 == 0xC0:
					L = 4
					R = rune(data[w-4]&0x07)<<18 |
						rune(data[w-3]&0x3F)<<12 |
						rune(data[w-2]&0x3F)<<6 |
						rune(data[w-1]&0x3F)
				}
				// Combining rune?
				if !unicode.Is(unicode.Me, R) && !unicode.Is(unicode.Mn, R) {
					comb = false
				}
				w -= L
				copy(data[w:w+L], Z)
			}
		case LoC1 <= R && R <= HiC1:
			// Drop C1 control codes
		default:
			// Write saved rune
			copy(data[w:], B[:L])
			w += L
		}

	}

	*in = (*in)[:w]
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

	// Say goodbye to client and reset default colors
	c.Write([]byte(text.Info + "\nBye bye...\n\n" + text.Reset))

	// Was the frontend closed?
	_, feClosed := c.Error().(frontend.ClosedError)

	switch {
	case c.Error() == nil:
		// No error - nothing to report
	case c.Error() == io.EOF:
		// io.EOF does not give address info so handle specially
		c.log("connection error: connection dropped by remote client")
	case feClosed:
		// Not an error so report without "Connection error:" prefix
		c.log("%s", c.Error())
	case !config.Server.LogClient:
		// If not logging client IP addresses make sure we don't leak them in any
		// error messages from the standard library
		e := c.Error().Error()
		e = strings.Replace(e, c.RemoteAddr().String(), "???", -1)
		c.log("connection error: %s", e)
	default:
		c.log("connection error: %s", c.Error())
	}

	// Make sure connection closed down and deallocated
	if err := c.Close(); err != nil {
		c.log("error closing connection: %s", err)
	} else {
		c.log("connection closed")
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

	c.SetWriteDeadline(time.Now().Add(writeTimeout))

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
