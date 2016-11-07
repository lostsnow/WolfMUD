// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package internal

import (
	"code.wolfmud.org/WolfMUD.git/has"

	"bytes"
)

// buffer is our extended version of a bytes.Buffer so that we can add some
// convience methods.
type buffer struct {
	*bytes.Buffer
}

type buffers map[has.Inventory]*buffer

// Msg is a collection of buffers for gathering messages to send back as a
// result of processing a command.
//
// NOTE: Observer is setup as an 'alias' for Observers[s.where] - Observer and
// Observers[s.where] point to the same buffer.
type Msg struct {
	Actor       *buffer
	Participant *buffer
	Observer    *buffer
	Observers   buffers
}

// WriteStrings takes a number of strings and writes them into the buffer. It's
// a convenience method to save writing multiple WriteString statements and an
// alternative to additional allocations due to concatenation.
//
// The return value n is the total length of all s, in bytes; err is always nil.
// The underlying bytes.Buffer may panic if it becomes too large.
func (b *buffer) WriteStrings(s ...string) (n int, err error) {
	for _, s := range s {
		x, _ := b.WriteString(s)
		n += x
	}
	return n, nil
}

// Allocate sets up the message buffers for the actor, participant and
// observers. The where passed in should be the current location so that
// Observer can be linked to the correct Observers element. The locks passed in
// are used to setup a buffer for observers in each location being locked.
//
// The participant and observers buffers need an initial linefeed to move the
// cursor off of the client's prompt line - for the actor this is done when
// they hit enter. The actor's buffer is initially set to half a page (half of
// 80 columns by 24 lines) as it is common to be sending location descriptions
// back to the actor. Half a page is arbitrary but seems to be reasonable.
func (m *Msg) Allocate(where has.Inventory, locks []has.Inventory) {
	if m.Actor == nil {
		m.Actor = &buffer{Buffer: bytes.NewBuffer(make([]byte, 0, (80*24)/2))}
		m.Participant = &buffer{Buffer: &bytes.Buffer{}}
		m.Observers = make(map[has.Inventory]*buffer)
		m.Participant.WriteByte(byte('\n'))
	}

	for _, l := range locks {
		if _, ok := m.Observers[l]; !ok {
			m.Observers[l] = &buffer{Buffer: &bytes.Buffer{}}
			m.Observers[l].WriteByte('\n')
		}
	}
	m.Observer = m.Observers[where]
}

// Deallocate releases the references to message buffers for the actor,
// participant and observers.
func (m *Msg) Deallocate() {
	m.Actor = nil
	m.Participant = nil
	m.Observer = nil
	for where := range m.Observers {
		m.Observers[where] = nil
		delete(m.Observers, where)
	}
}
