// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package internal

import (
	"bytes"
)

// buffer is our extended version of a bytes.Buffer so that we can add some
// convience methods.
type Buffer struct {
	*bytes.Buffer
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
