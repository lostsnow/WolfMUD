// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package internal

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"

	"bytes"
	"fmt"
	"strings"
	"testing"
)

// collectDelivery is a simple helper to deliver and collect messages. It uses
// a bytes.Buffer as a simple Write to deliver the messages, the content of
// which is returned.
func collectDelivery(b *buffer) (data []byte) {
	w := &bytes.Buffer{}
	b.Deliver(w)
	data = make([]byte, w.Len())
	w.Read(data)
	return
}

// newBuffers is a simple helper to populate buffers with a given number of
// buffer keyed by different Inventory.
func newBuffers(c int) (b buffers) {
	b = buffers{}
	for x := 0; x < c; x++ {
		b[attr.NewInventory()] = &buffer{}
	}
	return
}

// TestSimpleSend verifies basic message sending and delivery using a buffer.
func TestSimpleSend(t *testing.T) {

	// Some constants to make reading test cases easier
	const (
		silent = true
		noisy  = false

		noLF = true
		LF   = false
	)

	type message struct {
		data       []string
		omitLF     bool
		silentMode bool
	}

	testCases := []struct {
		messages []message // The messages being sent
		count    int       // Count of messages expected
		want     string    // Expected messages delivered
	}{
		// Single messages
		{
			[]message{
				{[]string{"Hello World!"}, noLF, silent},
			}, 0, "",
		},
		{
			[]message{
				{[]string{"Hello World!"}, noLF, noisy},
			}, 1, "Hello World!",
		},
		{
			[]message{
				{[]string{"Hello World!"}, LF, silent},
			}, 0, "",
		},
		{
			[]message{
				{[]string{"Hello World!"}, LF, noisy},
			}, 1, "\nHello World!",
		},

		// Two messages
		{
			[]message{
				{[]string{"Hello"}, noLF, silent},
				{[]string{"World!"}, noLF, silent},
			}, 0, "",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, silent},
				{[]string{"World!"}, noLF, noisy},
			}, 1, "World!",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, silent},
				{[]string{"World!"}, LF, silent},
			}, 0, "",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, silent},
				{[]string{"World!"}, LF, noisy},
			}, 1, "\nWorld!",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy},
				{[]string{"World!"}, noLF, silent},
			}, 1, "Hello",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy},
				{[]string{"World!"}, noLF, noisy},
			}, 2, "Hello\nWorld!",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy},
				{[]string{"World!"}, LF, silent},
			}, 1, "Hello",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy},
				{[]string{"World!"}, LF, noisy},
			}, 2, "Hello\nWorld!",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, silent},
				{[]string{"World!"}, noLF, silent},
			}, 0, "",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, silent},
				{[]string{"World!"}, noLF, noisy},
			}, 1, "World!",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, silent},
				{[]string{"World!"}, LF, silent},
			}, 0, "",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, silent},
				{[]string{"World!"}, LF, noisy},
			}, 1, "\nWorld!",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy},
				{[]string{"World!"}, noLF, silent},
			}, 1, "\nHello",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy},
				{[]string{"World!"}, noLF, noisy},
			}, 2, "\nHello\nWorld!",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy},
				{[]string{"World!"}, LF, silent},
			}, 1, "\nHello",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy},
				{[]string{"World!"}, LF, noisy},
			}, 2, "\nHello\nWorld!",
		},
	}

	for x, c := range testCases {
		t.Run(fmt.Sprintf("%d", x), func(t *testing.T) {

			b := &buffer{}

			// Send messages to the buffer with omitLF & silentMode flags
			for _, m := range c.messages {
				b.omitLF = m.omitLF
				b.Silent(m.silentMode)
				b.Send(m.data...)

				// Check silent return correct mode and mode set correctly
				want := m.silentMode
				have := b.Silent(true) // Get old mode
				if want != have {
					t.Errorf("SimpleSend incorrect silent mode have: %t want: %t", have, want)
				}
				b.Silent(have) // Reset to previous old mode
			}

			{ // Check number of messages in buffer
				want := c.count
				have := b.Len()
				if b.Len() != c.count {
					t.Errorf("SimpleSend wrong message count have: %d want: %d", have, want)
				}
			}

			{ // Check delivery of messages
				have := collectDelivery(b)

				if !bytes.Equal(have, []byte(c.want)) {
					t.Errorf("SimpleSend have: %+q want: %+q", have, c.want)
				}
			}
		})
	}
}

// TestSendAppend verifies message sending, appending and delivery using a buffer.
func TestSendAppend(t *testing.T) {

	// Some constants to make reading test cases easier
	const (
		silent = true
		noisy  = false

		noLF = true
		LF   = false

		append = true
		send   = false
	)

	type message struct {
		data       []string
		omitLF     bool
		silentMode bool
		append     bool
	}

	testCases := []struct {
		messages []message // The messages being sent
		count    int       // Count of messages expected
		want     string    // Expected messages delivered
	}{
		// Single messages
		{
			[]message{
				{[]string{"Hello World!"}, noLF, silent, append},
			}, 0, "",
		},
		{
			[]message{
				{[]string{"Hello World!"}, noLF, noisy, append},
			}, 1, "Hello World!",
		},
		{
			[]message{
				{[]string{"Hello World!"}, LF, silent, append},
			}, 0, "",
		},
		{
			[]message{
				{[]string{"Hello World!"}, LF, noisy, append},
			}, 1, "\nHello World!",
		},
		{
			[]message{
				{[]string{""}, LF, noisy, append},
			}, 1, "\n",
		},

		// Two messages
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, append},
				{[]string{"World!"}, noLF, noisy, append},
			}, 1, "Hello World!",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy, append},
				{[]string{"World!"}, LF, noisy, append},
			}, 1, "\nHello World!",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, send},
				{[]string{"World!"}, noLF, noisy, append},
			}, 1, "Hello World!",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy, send},
				{[]string{"World!"}, LF, noisy, append},
			}, 1, "\nHello World!",
		},

		// Three messages
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, send},
				{[]string{"World"}, noLF, noisy, append},
				{[]string{"!"}, noLF, noisy, append},
			}, 1, "Hello World !",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, append},
				{[]string{"World"}, noLF, noisy, append},
				{[]string{"!"}, noLF, noisy, append},
			}, 1, "Hello World !",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, append},
				{[]string{"World"}, noLF, silent, append},
				{[]string{"!"}, noLF, noisy, append},
			}, 1, "Hello !",
		},

		// Four messages
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, send},
				{[]string{"World!"}, noLF, noisy, append},
				{[]string{"Hello"}, noLF, noisy, send},
				{[]string{"World!"}, noLF, noisy, append},
			}, 2, "Hello World!\nHello World!",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy, send},
				{[]string{"World!"}, LF, noisy, append},
				{[]string{"Hello"}, LF, noisy, send},
				{[]string{"World!"}, LF, noisy, append},
			}, 2, "\nHello World!\nHello World!",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, silent, send},
				{[]string{"World!"}, LF, noisy, append},
				{[]string{"Hello"}, LF, noisy, send},
				{[]string{"World!"}, LF, noisy, append},
			}, 2, "\nWorld!\nHello World!",
		},

		// Location example
		{
			[]message{
				{[]string{"[ Somewhere ]"}, LF, noisy, send},
				{[]string{""}, LF, noisy, send},
				{[]string{"This is somewhere."}, LF, noisy, append},
			}, 2, "\n[ Somewhere ]\nThis is somewhere.",
		},
	}

	for x, c := range testCases {
		t.Run(fmt.Sprintf("%d", x), func(t *testing.T) {

			b := &buffer{}

			// Send messages to the buffer with omitLF & silentMode flags
			for _, m := range c.messages {
				b.omitLF = m.omitLF
				b.Silent(m.silentMode)
				if m.append {
					b.Append(m.data...)
				} else {
					b.Send(m.data...)
				}
			}

			{ // Check number of messages in buffer
				want := c.count
				have := b.Len()
				if b.Len() != c.count {
					t.Errorf("SendAppend wrong message count have: %d want: %d %+q", have, want, b.buf)
				}
			}

			{ // Check delivery of messages
				have := collectDelivery(b)

				if !bytes.Equal(have, []byte(c.want)) {
					t.Errorf("SendAppend have: %+q want: %+q", have, c.want)
				}
			}
		})
	}
}

// TestCombo tests all of the different combinations of omitLF and silent flags
// for Send and Append buffer methods sending/appending two messages to a
// buffer.
func TestCombo(t *testing.T) {

	m := []struct {
		first  []string
		second []string
	}{
		{
			[]string{"Hello", "World"},
			[]string{"Goodbye", "Universe"},
		},
		{
			[]string{"", ""},
			[]string{"", ""},
		},
	}

	for _, m := range m {

		// Loop through binary 000000 to 111111 for our six test flag combinations
		for x := 0; x < 1<<6; x++ {

			// Set up our six testing flags 3 each for 2 messages
			omitLF1 := x&32 != 0
			silent1 := x&16 != 0
			append1 := x&8 != 0
			omitLF2 := x&4 != 0
			silent2 := x&2 != 0
			append2 := x&1 != 0

			t.Run(fmt.Sprintf("Combo %t %t %t %t %t %t", omitLF1, silent1, append1, omitLF2, silent2, append2), func(t *testing.T) {

				b := &buffer{}

				// Send/append first message
				b.omitLF = omitLF1
				b.Silent(silent1)
				if append1 {
					b.Append(m.first...)
				} else {
					b.Send(m.first...)
				}

				// Send/append second message
				b.omitLF = omitLF2
				b.Silent(silent2)
				if append2 {
					b.Append(m.second...)
				} else {
					b.Send(m.second...)
				}

				// Manually build want based on current flags for first message
				want := ""
				count := 0
				if !silent1 {
					if len(want) == 0 {
						if !omitLF1 {
							want = want + "\n"
						}
						if append1 {
							count++
						}
					} else {
						if append1 {
							if want[len(want)-1] != '\n' {
								want = want + " "
							}
						} else {
							want = want + "\n"
						}
					}
					if append1 {
						want = want + strings.Join(m.first, "")
					} else {
						want = want + strings.Join(m.first, "")
						count++
					}
				}

				// Manually build want based on current flags for second message
				if !silent2 {
					if len(want) == 0 {
						if !omitLF2 {
							want = want + "\n"
						}
						if append2 {
							count++
						}
					} else {
						if append2 {
							if want[len(want)-1] != '\n' {
								want = want + " "
							}
						} else {
							want = want + "\n"
						}
					}
					if append2 {
						want = want + strings.Join(m.second, "")
					} else {
						want = want + strings.Join(m.second, "")
						count++
					}
				}

				{ // Check count
					have := b.Len()
					want := count
					if have != want {
						t.Errorf("Combo wrong message count: %d want: %d %+q", have, want, b.buf)
					}
				}

				{ // Check delivery of messages
					have := collectDelivery(b)

					if !bytes.Equal(have, []byte(want)) {
						t.Errorf("Combo have: %+q want: %+q", have, want)
					}
				}
			})
		}
	}
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
