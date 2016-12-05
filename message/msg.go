// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package message

import (
	"code.wolfmud.org/WolfMUD.git/has"
)

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
