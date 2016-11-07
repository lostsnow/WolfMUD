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
type Buffer struct {
	*bytes.Buffer
}

type Buffers map[has.Inventory]*Buffer

// Msg is a collection of buffers for gathering messages to send back as a
// result of processing a command.
//
// NOTE: Observer is setup as an 'alias' for Observers[s.where] - Observer and
// Observers[s.where] point to the same buffer.
type Msg struct {
	Actor       *Buffer
	Participant *Buffer
	Observer    *Buffer
	Observers   Buffers
}

// WriteStrings takes a number of strings and writes them into the buffer. It's
// a convenience method to save writing multiple WriteString statements and an
// alternative to additional allocations due to concatenation.
//
// The return value n is the total length of all s, in bytes; err is always nil.
// The underlying bytes.Buffer may panic if it becomes too large.
func (b *Buffer) WriteStrings(s ...string) (n int, err error) {
	for _, s := range s {
		x, _ := b.WriteString(s)
		n += x
	}
	return n, nil
}
