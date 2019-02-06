// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package message

import (
	"io"
	"log"
	"runtime"

	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Buffer provides temporary storage for messages to players. The Buffer
// accumulates messages which can then be sent as single network writes to the
// players. A Buffer can handle insertion of line feeds into messages
// automatically when required.
//
// While a *Buffer can be created using &Buffer{} it is a better to use calls
// to AcquireBuffer and ReleaseBuffer which will supply a *Buffer from a
// reusable pool of *Buffer and return the *Buffer to the pool for reuse.
//
// NOTE: omitLF indicates whether an empty Buffer should start with a line feed
// or not. This should be true for an actor's Buffer as they would have moved
// to a new line when pressing enter to issue a command. For all other Buffers
// it should be false as we need to move them off their prompt line manually.
type Buffer struct {
	buf        []byte
	omitLF     bool // Omit initial line feed?
	silentMode bool
	count      int // Number of messages in a Buffer
}

// pool is a collection of reusable *Buffer. A *Buffer can be obtained from the
// pool by calling AcquireBuffer and returned by calling ReleaseBuffer. The
// pool is large enough for four buffers - actor, participant and two observers
// (common case for moving between locations) - per 128 players, per CPU
// available. However the pool size is arbitrary and a typical compromise
// between space and performance. We don't need a huge pool as there are only
// so many goroutines that can be running on so many CPUs at any given time.
// Extra buffers will be allocated if needed and dropped again when the pool is
// full.
var pool = make(
	chan *Buffer,
	(int)(config.Server.MaxPlayers/128)*4*runtime.GOMAXPROCS(-1),
)

// init reports the size of the *Buffer pool.
func init() {
	log.Printf("Allocated pool for %d buffers", cap(pool))
}

// AcquireBuffer returns a *Buffer from a  pool of *Buffer. A *Buffer should be
// returned to the pool by calling ReleaseBuffer. It is not essential that
// ReleaseBuffer is called as the pool will replenish itself, however a *Buffer
// that is simply discarded cannot be reused - which avoids allocations and
// generating garbage.
func AcquireBuffer() (b *Buffer) {
	select {
	case b = <-pool:
		// Make sure buffer is reset
		b.buf = b.buf[0:0]
		b.omitLF = false
		b.silentMode = false
		b.count = 0
	default:
		b = &Buffer{}
	}
	return
}

// ReleaseBuffer puts a *Buffer back into a pool of *Buffer for reuse.
func ReleaseBuffer(b *Buffer) {
	if b == nil {
		return
	}
	select {
	case pool <- b:
	default:
	}
}

// Send takes a number of strings and writes them into the Buffer as a single
// message. The message will automatically be prefixed with a line feed if
// required so that the message starts on its own new line when displayed to
// the player. Each time Send is called the message count returned by Len is
// increased by one.
//
// If the Buffer is in silent mode the Buffer and message count will not be
// modified and the passed strings will be discarded.
func (b *Buffer) Send(s ...string) {
	if b == nil || b.silentMode {
		return
	}
	if b.count != 0 || !b.omitLF {
		b.buf = append(b.buf, '\n')
	}
	for _, s := range s {
		b.buf = append(b.buf, s...)
	}
	b.count++
	return
}

// sendColor is the same as Send but it also writes a color string such as
// text.Bad or text.Red before the given strings of the message.
//
// The code of this method is copied from Send to avoid allocations prefixing
// the color string to the strings of the message and then calling Send.
func (b *Buffer) sendColor(c string, s ...string) {
	if b == nil || b.silentMode {
		return
	}
	if b.count != 0 || !b.omitLF {
		b.buf = append(b.buf, '\n')
	}
	b.buf = append(b.buf, c...)
	for _, s := range s {
		b.buf = append(b.buf, s...)
	}
	b.count++
	return
}

// SendGood is convenient for sending a message using text.Good as the color.
func (b *Buffer) SendGood(s ...string) { b.sendColor(text.Good, s...) }

// SendBad is convenient for sending a message using text.Bad as the color.
func (b *Buffer) SendBad(s ...string) { b.sendColor(text.Bad, s...) }

// SendInfo is convenient for sending a message using text.Info as the color.
func (b *Buffer) SendInfo(s ...string) { b.sendColor(text.Info, s...) }

// Append takes a number of strings and writes them into the Buffer appending
// to a previous message. The message is appended to the current Buffer with a
// leading single space. Append is useful when a message needs to be composed
// in several stages. Append does not normally increase the message count
// returned by Len, but see special cases below.
//
// If the Buffer is in silent mode the Buffer will not be modified and the
// passed strings will be discarded.
//
// Special cases:
//
// If Append is called without an initial Send then Append will behave like a
// Send and also increase the message count by one.
//
// If Append is called without an initial Send or after a Send with an empty
// string the leading space will be omitted. This is so that Send can cause the
// start a new message but text is only appended by calling Append.
func (b *Buffer) Append(s ...string) {
	if b == nil || b.silentMode {
		return
	}

	// If Buffer is empty we have to start a new message, otherwise append with a
	// single space
	if b.count == 0 {
		if !b.omitLF {
			b.buf = append(b.buf, '\n')
		}
		b.count++
	} else {
		// We don't append a space right after a line feed
		if l := len(b.buf); l != 0 && b.buf[l-1] != '\n' {
			b.buf = append(b.buf, ' ')
		}
	}

	for _, s := range s {
		b.buf = append(b.buf, s...)
	}
	return
}

// Silent sets a Buffer silent mode to true or false and returning the old
// silent mode. When a Buffer is in silent mode it will ignore calls to Send
// and Append.
func (b *Buffer) Silent(new bool) (old bool) {
	old, b.silentMode = b.silentMode, new
	return
}

// OmitLF sets a Buffer omitLF flag to true or false and returns the old omitLF
// setting. For details of the omitLF flag see the Buffer type.
func (b *Buffer) OmitLF(new bool) (old bool) {
	old, b.omitLF = b.omitLF, new
	return
}

// Len returns the number of messages in a Buffer.
func (b *Buffer) Len() int {
	return b.count
}

var resetLen = len(text.Reset)

// Deliver writes all of the messages in the Buffer to the passed Writers.
// After the messages have been delivered the messages and message count will
// be cleared.
func (b *Buffer) Deliver(w ...io.Writer) {

	// If there are no messages and Buffer isn't the actor's make sure the Buffer
	// is cleared and just bail. For the actor there may be no messages if e.g.
	// they just hit enter, we still need to deliver nothing to write out a new
	// prompt for them.
	if b.count == 0 && !b.omitLF {
		b.buf = b.buf[0:0]
		return
	}

	// If Buffer does not start with an escape sequence insert a reset to
	// default colors
	if len(b.buf) > 0 && b.buf[0] != '\x1b' {
		b.buf = append(b.buf, text.Reset...)
		copy(b.buf[resetLen:], b.buf[0:len(b.buf)-resetLen])
		copy(b.buf[0:resetLen], text.Reset)
	}

	// Make sure prompt appears at start of next new line
	if b.count != 0 || !b.omitLF {
		b.buf = append(b.buf, '\n')
	}

	// If sending messages to a single writer don't make a copy
	if len(w) == 1 {
		w[0].Write(b.buf)
	}

	// If we have multiple writers write a copy of the Buffer to each
	if len(w) > 1 {
		for _, w := range w {
			c := make([]byte, len(b.buf))
			copy(c, b.buf)
			w.Write(c)
		}
	}

	// Reset Buffer for reuse
	b.buf = b.buf[0:0]
	b.count = 0
}
