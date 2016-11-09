// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package internal

import (
	"code.wolfmud.org/WolfMUD.git/has"
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
// the player. If the buffer is in silent mode the buffer will not be modified
// and the passed strings will be discarded.
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
	return
}

// Append takes a number of strings and write them into the buffer appending to
// a previous message. The message is appended to the current buffer with a
// single space prefixing it. This is useful when a message needs to be
// composed in several stages. It is safe to call Append without having first
// called Send - this will cause the first Append to act like an initial Send.
// If the buffer is in silent mode the buffer will not be modified and the
// passed strings will be discarded.
func (b *buffer) Append(s ...string) {
	if b.silentMode {
		return
	}
	if len(b.buf) == 0 && !b.omitLF {
		b.buf = append(b.buf, '\n')
	}
	if l := len(b.buf); l != 0 && b.buf[l-1] != '\n' {
		b.buf = append(b.buf, ' ')
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

// Send calls buffer.Send for each buffer in the receiver buffers.
//
// See buffer.Send for more details.
func (b buffers) Send(s ...string) {
	for _, b := range b {
		b.Send(s...)
	}
}

// Append calls buffer.Append for each buffer in the receiver buffers.
//
// See buffer.Append for more details.
func (b buffers) Append(s ...string) {
	for _, b := range b {
		b.Append(s...)
	}
}

// Silent calls buffer.Silent for each buffer in the receiver buffers. If there
// are less new state flags passed in than there are buffer in buffers the last
// flag will be repeated for each remaining buffe. Silent returns the old
// silent state of each buffer as a []bool.
//
// See buffer.Silent for more details.
func (b buffers) Silent(new ...bool) (old []bool) {

	// Make sure we have enough new flags for each buffer by repeating the last
	// new flag if required
	lastNew := new[len(new)-1]
	for x := len(b) - len(new); x > 0; x-- {
		new = append(new, lastNew)
	}

	// Assign each new flag to each buffer
	x := 0
	for _, b := range b {
		old = append(old, b.Silent(new[x]))
		x++
	}
	return
}

// Temporary methods to facilitate switch from bytes.Buffer to []bytes
func (b *buffer) Len() int       { return len(b.buf) }
func (b *buffer) Bytes() []byte  { return b.buf }
func (b *buffer) Truncate(l int) { b.buf = b.buf[:l] }

// Allocate sets up the message buffers for the actor, participant and
// observers. The 'where' passed in should be the current location so that
// Observer can be linked to the correct Observers element. The locks passed in
// are used to setup a buffer for observers in each of the locations being
// locked.
//
// The actor's buffer is initially set to half a page (half of 80 columns by 24
// lines) as it is common to be sending location descriptions back to the
// actor. Half a page is arbitrary but seems to be reasonable.
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
