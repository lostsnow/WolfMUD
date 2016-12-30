// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package message

import (
	"code.wolfmud.org/WolfMUD.git/text"

	"io"
)

// buffer provides temporary storeage for messages to players. The buffer
// accumulates messages which can then be sent as single network writes to the
// players. A buffer can handle insertion of line feeds into messages
// automatically when required.
//
// NOTE: omitLF indicates whether an empty buffer should start with a line feed
// or not. This should be true for an actor's buffer as they would have moved
// to a new line when pressing enter to issue a command. For all other buffers
// it should be false as we need to move them off their prompt line manually.
type buffer struct {
	buf        []byte
	omitLF     bool // Omit initial line feed?
	silentMode bool
	count      int // Number of messages in a buffer
}

// Buffer allows a buffer to be embedded in a struct without exposing buffer
// itself. A buffer can be created and assigned using NewBuffer.
type Buffer interface {
	Send(...string)
	Append(...string)
	Silent(bool) bool
	Len() int
	Deliver(w ...io.Writer)
}

// NewBuffer returns a buffer with omitLF set to true - suitable for use as a
// standalone buffer.
func NewBuffer() (b *buffer) {
	b = &buffer{}
	b.omitLF = true
	return
}

// Send takes a number of strings and writes them into the buffer as a single
// message. The message will automatically be prefixed with a line feed if
// required so that the message starts on its own new line when displayed to
// the player. Each time Send is called the message count returned by Len is
// increased by one.
//
// If the buffer is in silent mode the buffer and message count will not be
// modified and the passed strings will be discarded.
func (b *buffer) Send(s ...string) {
	if b.silentMode {
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
func (b *buffer) sendColor(c string, s ...string) {
	if b.silentMode {
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
func (b *buffer) SendGood(s ...string) { b.sendColor(text.Good, s...) }

// SendBad is convenient for sending a message using text.Bad as the color.
func (b *buffer) SendBad(s ...string) { b.sendColor(text.Bad, s...) }

// SendInfo is convenient for sending a message using text.Info as the color.
func (b *buffer) SendInfo(s ...string) { b.sendColor(text.Info, s...) }

// Append takes a number of strings and writes them into the buffer appending
// to a previous message. The message is appended to the current buffer with a
// leading single space. Append is useful when a message needs to be composed
// in several stages. Append does not normally increase the message count
// returned by Len, but see special cases below.
//
// If the buffer is in silent mode the buffer will not be modified and the
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
func (b *buffer) Append(s ...string) {
	if b.silentMode {
		return
	}

	// If buffer is empty we have to start a new message, otherwise append with a
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

// Silent sets a buffers silent mode to true or false and returning the old
// silent mode. When a buffer is in silent mode it will ignore calls to Send
// and Append.
func (b *buffer) Silent(new bool) (old bool) {
	old, b.silentMode = b.silentMode, new
	return
}

// Len returns the number of messages in a buffer.
func (b *buffer) Len() int {
	return b.count
}

var resetLen = len(text.Reset)

// Deliver writes all of the messages in the buffer to the passed Writers.
// After the messages have been delivered the messages and message count will
// be cleared.
func (b *buffer) Deliver(w ...io.Writer) {

	// If there are no messages and buffer isn't the actor's make sure the buffer
	// is cleared and just bail. For the actor there may be no messages if e.g.
	// they just hit enter, we still need to deliver nothing to write out a new
	// prompt for them.
	if b.count == 0 && !b.omitLF {
		b.buf = b.buf[0:0]
		return
	}

	// If buffer does not start with an escape sequence insert a reset to
	// default colors
	if len(b.buf) > 0 && b.buf[0] != '\033' {
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

	// If we have multiple writers write a copy of the buffer to each
	if len(w) > 1 {
		for _, w := range w {
			c := make([]byte, len(b.buf))
			copy(c, b.buf)
			w.Write(c)
		}
	}

	// Clear down the messages and message count
	b.buf = b.buf[0:0]
	b.count = 0
}
