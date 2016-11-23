// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package message

import (
	"code.wolfmud.org/WolfMUD.git/has"

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
	Deliver(w io.Writer)
}

// NewBuffer returns a buffer with omitLF set to true - suitable for use as a
// standalone buffer.
func NewBuffer() (b *buffer) {
	b = &buffer{}
	b.omitLF = true
	return
}

// buffers are a collection of buffer indexed by location.
type buffers map[has.Inventory]*buffer

// Msg is a collection of buffers for gathering messages to send back as a
// result of processing a command. Before use a Msg should have Allocate called
// on it to allocate and setup the buffers internally. After use Deallocate
// should be called to free up the buffers. The Allocate and Deallocate methods
// are kept separate so that a Msg can be reused by repeated calls to
// Allocate/Deallocate.
//
// NOTE: Observer is setup as an 'alias' for Observers[s.where] - Observer and
// Observers[s.where] point to the same buffer. See the Allocate method for
// more details.
type Msg struct {
	Actor       *buffer
	Participant *buffer
	Observer    *buffer
	Observers   buffers
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
	if len(b.buf) != 0 || !b.omitLF {
		b.buf = append(b.buf, '\n')
	}
	for _, s := range s {
		b.buf = append(b.buf, s...)
	}
	b.count++
	return
}

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
	if l := len(b.buf); l == 0 {
		if !b.omitLF {
			b.buf = append(b.buf, '\n')
		}
		b.count++
	} else {
		// We don't append a space right after a line feed
		if b.buf[l-1] != '\n' {
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

// Deliver writes all of the messages in the deliver buffer to the given
// Writer.
func (b *buffer) Deliver(w io.Writer) {

	// If there are no messages and buffer isn't the actor's just bail.
	// For the actor there may be no messages if e.g. they just hit enter, we
	// still need to deliver nothing to write out a new prompt for them.
	if b.count == 0 && !b.omitLF {
		return
	}

	w.Write(b.buf)

	// Only clear down buffers with omitLF set for reuse as these are intended to
	// be standalone and reusable.
	if b.omitLF {
		b.buf = b.buf[0:0]
		b.count = 0
	}
}

// Send calls buffer.Send for each buffer in the receiver buffers.
//
// See also buffer.Send for more details.
func (b buffers) Send(s ...string) {
	for _, b := range b {
		b.Send(s...)
	}
}

// Append calls buffer.Append for each buffer in the receiver buffers.
//
// See also buffer.Append for more details.
func (b buffers) Append(s ...string) {
	for _, b := range b {
		b.Append(s...)
	}
}

// Silent calls buffer.Silent with the passed new flag for each buffer in the
// receiver buffers. Silent returns two sets of buffers, one for all buffers
// that were true and one for all buffers that were false. The previous silent
// state of buffers can be restored by calling Silent with true or false on the
// returned buffers. For example:
//
//	t,f := s.msg.Observers.Silent(true)
//	:
//	: // do something
//	:
//	t.Silent(true)
//	f.silent(false)
//
// See also buffer.Silent for more details.
func (b buffers) Silent(new bool) (t buffers, f buffers) {
	t = make(map[has.Inventory]*buffer)
	f = make(map[has.Inventory]*buffer)
	for where, b := range b {
		if old := b.Silent(new); old {
			t[where] = b
		} else {
			f[where] = b
		}
	}
	return
}

// Len returns the number of messages for each buffer in buffers as a
// [has.Inventory]int map.
func (b buffers) Len() (l map[has.Inventory]int) {
	l = make(map[has.Inventory]int)
	for where, b := range b {
		l[where] = b.count
	}
	return
}

// Filter takes a list of Inventories and returns only matching buffer entries
// as buffers.
func (b buffers) Filter(limit ...has.Inventory) (filtered buffers) {
	filtered = make(map[has.Inventory]*buffer)
	for _, l := range limit {
		if _, ok := b[l]; ok {
			filtered[l] = b[l]
		}
	}
	return
}

// Allocate sets up the message buffers for the actor, participant and
// observers. The 'where' passed in should be the current location so that
// Observer can be linked to the correct Observers element. The locks passed in
// are used to setup a buffer for observers in each of the locations being
// locked.
//
// The actor's buffer is initially set to half a page (half of 80 columns by 24
// lines) as it is common to be sending location descriptions back to the
// actor. Half a page is arbitrary but seems to be reasonable.
//
// NOTE: For crowded locations buffers of observers are automatically put in
// silent mode.
func (m *Msg) Allocate(where has.Inventory, locks []has.Inventory) {
	if m.Actor == nil {
		m.Actor = &buffer{buf: make([]byte, 0, (80*24)/2)}
		m.Actor.omitLF = true
		m.Participant = &buffer{}
		m.Observers = make(map[has.Inventory]*buffer)
	}

	for _, l := range locks {
		if _, ok := m.Observers[l]; !ok {
			m.Observers[l] = &buffer{}
			m.Observers[l].Silent(l.Crowded())
		}
	}
	m.Observer = m.Observers[where]
}

// Deallocate releases the references to message buffers for the actor,
// participant and observers. Specific deallocation can help with garbage
// collection.
func (m *Msg) Deallocate() {
	m.Actor = nil
	m.Participant = nil
	m.Observer = nil
	for where := range m.Observers {
		m.Observers[where] = nil
		delete(m.Observers, where)
	}
}
