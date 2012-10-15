// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package command

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	. "code.wolfmud.org/WolfMUD.git/utils/test"
)

// Define a command issuer harness for testing. This is a minimal Thing that
// implements messaging.Responder and messaging.Broadcaster that captures the
// messages sent so we can compare what was received with what we expected. We
// can also use it to issue test commands as it is a Thing.
type testHarness struct {
	thing.Thing
	response  string
	broadcast string
}

func (h *testHarness) String() string {
	return strconv.Itoa((int)(h.UniqueId()))
}

func NewTestHarness() *testHarness {
	return &testHarness{
		Thing: *thing.New("Issuer", []string{"ISSUER"}, "A test issuer"),
	}
}

func (h *testHarness) Respond(format string, any ...interface{}) {
	h.response += fmt.Sprintf(format, any...)
}

func (h *testHarness) Broadcast(omit []thing.Interface, format string, any ...interface{}) {
	for _, omit := range omit {
		if omit.IsAlso(h) {
			return
		}
	}
	h.broadcast += fmt.Sprintf(format, any...)
}

func (h *testHarness) clearBuffers() {
	h.response, h.broadcast = "", ""
}

type testSubject struct {
	cmd   string   // Command to issue
	verb  string   // Expected verb from issued command
	nouns []string // Expected nouns from issued command
}

var testSubjects = []testSubject{
	{"foo", "FOO", []string{}},
	{"bar ball", "BAR", []string{"BALL"}},
	{"foo ball lattice", "FOO", []string{"BALL", "LATTICE"}},
	{"bar ball lattice", "BAR", []string{"BALL", "LATTICE"}},
	{"foo ball", "FOO", []string{"BALL"}},
	{"bar", "BAR", []string{}},
}

func checkCommandStruct(t *testing.T, h *testHarness, s testSubject, c *Command) {
	Equal(t, "New issuer", h.UniqueId(), c.Issuer.UniqueId())
	Equal(t, "New verb", s.verb, c.Verb)
	Equal(t, "New noun", len(s.nouns), len(c.Nouns))
	for i, n := range s.nouns {
		Equal(t, "New noun", n, c.Nouns[i])
	}
	if len(s.nouns) > 0 {
		Equal(t, "New target", s.nouns[0], c.Target)
	} else {
		Equal(t, "New target", "", c.Target)
	}
}

func TestFuncNew(t *testing.T) {
	h := NewTestHarness()

	for _, s := range testSubjects {
		checkCommandStruct(t, h, s, New(h, s.cmd))
	}
}

func TestMethodNew(t *testing.T) {
	h := NewTestHarness()
	cmd := New(h, "")

	for _, s := range testSubjects {
		cmd.New(s.cmd)
		checkCommandStruct(t, h, s, cmd)
	}
}

var testMessages = [][]string{
	{"Single message test"},
	{"Hello World!", "How are you?"},
	{""},
	{"", ""},
	{"This is", "another multi-line", "test - but now", "with extra added", "lines and vitamin caffine ;)"},
}

func TestRespond(t *testing.T) {
	for _, messages := range testMessages {
		h := NewTestHarness()
		cmd := New(h, "")
		for _, msg := range messages {
			cmd.Respond(msg)
		}
		cmd.Flush()
		Equal(t, "Respond", strings.Join(messages, "\n"), h.response)
	}
}

// Make sure flush is working and clearing the buffers
func TestRespondFlush(t *testing.T) {
	h := NewTestHarness()
	cmd := New(h, "")
	for _, messages := range testMessages {
		for _, msg := range messages {
			cmd.Respond(msg)
		}
		cmd.Flush()
		Equal(t, "Respond Flush", strings.Join(messages, "\n"), h.response)
		h.clearBuffers()
	}
}

func TestBroadcast(t *testing.T) {
	for _, messages := range testMessages {
		h := NewTestHarness()
		cmd := New(h, "")
		for _, msg := range messages {
			cmd.Broadcast(nil, msg)
		}
		cmd.Flush()
		Equal(t, "Broadcast", strings.Join(messages, "\n"), h.broadcast)
	}
}

// Make sure flush is working and clearing the buffers
func TestBroadcastFlush(t *testing.T) {
	h := NewTestHarness()
	cmd := New(h, "")
	for _, messages := range testMessages {
		for _, msg := range messages {
			cmd.Broadcast(nil, msg)
		}
		cmd.Flush()
		Equal(t, "Broadcast Flush", strings.Join(messages, "\n"), h.broadcast)
		h.clearBuffers()
	}
}

func TestBroadcastOmit(t *testing.T) {
	for _, messages := range testMessages {
		h := NewTestHarness()
		omit := []thing.Interface{h}
		cmd := New(h, "")
		for _, msg := range messages {
			cmd.Broadcast(omit, msg)
		}
		cmd.Flush()
		Equal(t, "Broadcast (omit)", "", h.broadcast)
	}
}

// The main locking functions are: AddLock,	LocksModified and CanLock which are
// difficult to test on their own. So TestLocking tests all of them together.
func TestLocking(t *testing.T) {

	things := make([]thing.Interface, 10, 10)
	for x, _ := range things {
		things[x] = NewTestHarness()
	}

	// Try tests twice. 1st time with things slice as created. 2nd time with
	// things slice reversed. This tests the AddLock ordering.
	for try := 1; try < 3; try++ {
		cmd := New(things[0], "")

		for x, h := range things {
			Equal(t, "Locks length (1)", x, len(cmd.Locks))

			// Check twice as LocksModified() resets when called
			Equal(t, "Locks modified (1)", false, cmd.LocksModified())
			Equal(t, "Locks modified (2)", false, cmd.LocksModified())

			for y, h := range things {
				Equal(t, "Locks can lock (1)", y < x, cmd.CanLock(h))
			}

			cmd.AddLock(h)
			Equal(t, "Locks length (2)", x+1, len(cmd.Locks))

			// Check twice as LocksModified() resets when called
			Equal(t, "Locks modified (3)", true, cmd.LocksModified())
			Equal(t, "Locks modified (4)", false, cmd.LocksModified())

			for y, h := range things {
				Equal(t, "Locks can lock (2)", y <= x, cmd.CanLock(h))
			}
		}

		// Reverse things slice for 2nd try - inplace without new allocations
		l := len(things)-1
		for x := (int)(l / 2); x >= 0; x-- {
			things[x], things[l-x] = things[l-x], things[x]
		}
	}
}
