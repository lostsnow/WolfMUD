// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package message

import (
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
			}, 1, "Hello World!\n",
		},
		{
			[]message{
				{[]string{"Hello World!"}, LF, silent},
			}, 0, "",
		},
		{
			[]message{
				{[]string{"Hello World!"}, LF, noisy},
			}, 1, "\nHello World!\n",
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
			}, 1, "World!\n",
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
			}, 1, "\nWorld!\n",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy},
				{[]string{"World!"}, noLF, silent},
			}, 1, "Hello\n",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy},
				{[]string{"World!"}, noLF, noisy},
			}, 2, "Hello\nWorld!\n",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy},
				{[]string{"World!"}, LF, silent},
			}, 1, "Hello\n",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy},
				{[]string{"World!"}, LF, noisy},
			}, 2, "Hello\nWorld!\n",
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
			}, 1, "World!\n",
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
			}, 1, "\nWorld!\n",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy},
				{[]string{"World!"}, noLF, silent},
			}, 1, "\nHello\n",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy},
				{[]string{"World!"}, noLF, noisy},
			}, 2, "\nHello\nWorld!\n",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy},
				{[]string{"World!"}, LF, silent},
			}, 1, "\nHello\n",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy},
				{[]string{"World!"}, LF, noisy},
			}, 2, "\nHello\nWorld!\n",
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
			}, 1, "Hello World!\n",
		},
		{
			[]message{
				{[]string{"Hello World!"}, LF, silent, append},
			}, 0, "",
		},
		{
			[]message{
				{[]string{"Hello World!"}, LF, noisy, append},
			}, 1, "\nHello World!\n",
		},
		{
			[]message{
				{[]string{""}, LF, noisy, append},
			}, 1, "\n\n",
		},

		// Two messages
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, append},
				{[]string{"World!"}, noLF, noisy, append},
			}, 1, "Hello World!\n",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy, append},
				{[]string{"World!"}, LF, noisy, append},
			}, 1, "\nHello World!\n",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, send},
				{[]string{"World!"}, noLF, noisy, append},
			}, 1, "Hello World!\n",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy, send},
				{[]string{"World!"}, LF, noisy, append},
			}, 1, "\nHello World!\n",
		},

		// Three messages
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, send},
				{[]string{"World"}, noLF, noisy, append},
				{[]string{"!"}, noLF, noisy, append},
			}, 1, "Hello World !\n",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, append},
				{[]string{"World"}, noLF, noisy, append},
				{[]string{"!"}, noLF, noisy, append},
			}, 1, "Hello World !\n",
		},
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, append},
				{[]string{"World"}, noLF, silent, append},
				{[]string{"!"}, noLF, noisy, append},
			}, 1, "Hello !\n",
		},

		// Four messages
		{
			[]message{
				{[]string{"Hello"}, noLF, noisy, send},
				{[]string{"World!"}, noLF, noisy, append},
				{[]string{"Hello"}, noLF, noisy, send},
				{[]string{"World!"}, noLF, noisy, append},
			}, 2, "Hello World!\nHello World!\n",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, noisy, send},
				{[]string{"World!"}, LF, noisy, append},
				{[]string{"Hello"}, LF, noisy, send},
				{[]string{"World!"}, LF, noisy, append},
			}, 2, "\nHello World!\nHello World!\n",
		},
		{
			[]message{
				{[]string{"Hello"}, LF, silent, send},
				{[]string{"World!"}, LF, noisy, append},
				{[]string{"Hello"}, LF, noisy, send},
				{[]string{"World!"}, LF, noisy, append},
			}, 2, "\nWorld!\nHello World!\n",
		},

		// Location example
		{
			[]message{
				{[]string{"[ Somewhere ]"}, LF, noisy, send},
				{[]string{""}, LF, noisy, send},
				{[]string{"This is somewhere."}, LF, noisy, append},
			}, 2, "\n[ Somewhere ]\nThis is somewhere.\n",
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

			t.Run(fmt.Sprintf("Combo o:%t s:%t a:%t o:%t s:%t a:%t", omitLF1, silent1, append1, omitLF2, silent2, append2), func(t *testing.T) {

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
					if count == 0 {
						if !omitLF1 {
							want = want + "\n"
						}
						if append1 {
							count++
						}
					} else {
						if append1 {
							if len(want) != 0 && want[len(want)-1] != '\n' {
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
					if count == 0 {
						if !omitLF2 {
							want = want + "\n"
						}
						if append2 {
							count++
						}
					} else {
						if append2 {
							if len(want) != 0 && want[len(want)-1] != '\n' {
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
				if count == 0 && !omitLF2 {
				} else {
					if count != 0 || !omitLF2 {
						want = want + "\n"
					}
				}

				{ // Check count
					havec := b.Len()
					wantc := count
					if havec != wantc {
						t.Errorf("Combo wrong message count: %d want: %d - %+q %+q", havec, wantc, b.buf, want)
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
