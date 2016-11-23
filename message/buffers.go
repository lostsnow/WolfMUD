// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package message

import (
	"code.wolfmud.org/WolfMUD.git/has"
)

// buffers are a collection of buffer indexed by location.
type buffers map[has.Inventory]*buffer

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
