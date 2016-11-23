// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package message

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"

	"bytes"
	"testing"
)

// newBuffers is a simple helper to populate buffers with a given number of
// buffer keyed by different Inventory.
func newBuffers(c int) (b buffers) {
	b = buffers{}
	for x := 0; x < c; x++ {
		b[attr.NewInventory()] = &buffer{}
	}
	return
}

// TestGroupSend tests sending messages to a group of buffer.
func TestGroupSend(t *testing.T) {

	// Use from 0 to 9 buffers
	for c := 0; c < 10; c++ {

		bufs := newBuffers(c)
		bufs.Send("Hello World!")

		// Check delivery of messages
		for _, b := range bufs {
			{
				have := collectDelivery(b)
				want := "\nHello World!"

				if !bytes.Equal(have, []byte(want)) {
					t.Errorf("GroupSend %d buffers have: %+q want: %+q", c, have, want)
				}
			}
		}
	}
}

// TestGroupSendAppend tests sending and appending messages to a group of
// buffer.
func TestGroupSendAppend(t *testing.T) {

	// Use from 0 to 9 buffers
	for c := 0; c < 10; c++ {

		bufs := newBuffers(c)
		bufs.Send("Hello")
		bufs.Append("World!")

		// Check delivery of messages
		for _, b := range bufs {
			{
				have := collectDelivery(b)
				want := "\nHello World!"

				if !bytes.Equal(have, []byte(want)) {
					t.Errorf("GroupSendAppend %d buffers have: %+q want: %+q", c, have, want)
				}
			}
		}
	}
}

// TestGroupSilent tests setting silent mode on a group of buffer.
func TestGroupSilent(t *testing.T) {

	// Use from 0 to 9 buffers
	for c := 0; c < 10; c++ {

		bufs := newBuffers(c)
		bufs.Silent(true)
		bufs.Send("Hello World!")

		// Check delivery of messages
		for _, b := range bufs {
			{
				have := collectDelivery(b)
				want := ""

				if !bytes.Equal(have, []byte(want)) {
					t.Errorf("GroupSilent buffer have: %+q want: %+q", have, want)
				}
			}
		}
	}
}

// TestGroupSilentReturn tests the state returned by Silent on a group of
// buffer.
func TestGroupSilentReturn(t *testing.T) {

	// Use from 0 to 9 buffers
	for c := uint(0); c < 10; c++ {

		// Setup silent pattern, 1 bit per buffer, going through all patterns
		for s := uint(0); s <= (1<<c)-1; s++ {

			// Maps for which inventories we want to be true (wt) and false (wf)
			wt := make(map[has.Inventory]struct{})
			wf := make(map[has.Inventory]struct{})

			// Setup buffers
			bufs := buffers{}
			for x := uint(0); x < c; x++ {
				b := &buffer{}
				where := attr.NewInventory()
				bufs[where] = b

				// Set silent according to bit set in s pattern and store inventory in
				// matching wanted true/false lists
				if s&(1<<x) != 0 {
					b.Silent(true)
					wt[where] = struct{}{}
				} else {
					wf[where] = struct{}{}
				}
			}

			// Set all buffers silent to retrieve previous state as have true (ht)
			// and have false (hf) lists
			ht, hf := bufs.Silent(true)

			// Check we have expected buffer in true lists
			for where := range wt {
				if _, ok := ht[where]; !ok {
					t.Errorf("GroupSilentReturn buffer not in true list")
				}
			}

			// Check we have expected buffer in false lists
			for where := range wf {
				if _, ok := hf[where]; !ok {
					t.Errorf("GroupSilentReturn buffer not in false list")
				}
			}
		}
	}
}

// TestGroupLen tests the message count returned by a group of buffer.
func TestGroupLen(t *testing.T) {
	bufs := newBuffers(10)
	want := make(map[has.Inventory]int)

	// Send a different, incrementing number of messages to each buffer. Record
	// number sent to each buffer in want list.
	x := 0
	for where, buf := range bufs {
		want[where] = x
		for y := 0; y < x; y++ {
			buf.Send("Hello World!")
		}
		x++
	}

	// Get number of messages in each buffer
	have := bufs.Len()

	// Make sure we have expected number of messages in each buffer
	for where, l := range have {
		if want[where] != l {
			t.Errorf("GroupLen wrong message count have: %d want: %d", l, want[where])
		}
	}
}

// TestGroupFilter tests that the correct buffer are returned by Filter.
func TestGroupFilter(t *testing.T) {
	bufs := newBuffers(10)
	want := []has.Inventory{}

	for where := range bufs {

		// Increase want list one buffer at a time and gradually filter buffers
		want = append(want, where)
		have := bufs.Filter(want...)

		// Check we have the expected number of buffer returned by Filter
		if lh, lw := len(have), len(want); lh != lw {
			t.Errorf("GroupFilter wrong number of buffers have: %d want: %d", lh, lw)
		}

		// Check buffers returned are the ones we wanted
		for _, where := range want {
			if _, ok := have[where]; !ok {
				t.Errorf("GroupFilter wrong buffer found")
			}
		}
	}
}
