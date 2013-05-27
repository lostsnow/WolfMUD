// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package command

import (
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"
	"code.wolfmud.org/WolfMUD.git/utils/uid"
	"fmt"
	"strings"
	"testing"
)

// Define simple mock Responder / Broadcaster that captures messages
type mock struct {
	thing.Thing
	ResponseBuf  string
	BroadcastBuf string
}

func newMock() (m *mock) {
	m = &mock{
		thing.Thing{},
		"",
		"",
	}
	m.Thing.Unmarshal(recordjar.Record{
		"name":    "Mock",
		"aliases": "MOCK",
		":data:":  "A mock",
	})
	return m
}

func (m *mock) Reset() { m.ResponseBuf, m.BroadcastBuf = "", "" }

func (m *mock) Respond(format string, any ...interface{}) {
	m.ResponseBuf += fmt.Sprintf(format, any...)
}

func (m *mock) Broadcast(omit []thing.Interface, format string, any ...interface{}) {
	for _, omit := range omit {
		if omit.IsAlso(m) {
			return
		}
	}
	m.BroadcastBuf += fmt.Sprintf(format, any...)
}

// END OF MOCK

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

func checkCommandStruct(t *testing.T, m *mock, s testSubject, c *Command) {

	// Check command is using right issuer
	{
		have := c.Issuer.UniqueId()
		want := m.UniqueId()
		if have != want {
			t.Errorf("Invalid unique ID: have %d wanted %d", have, want)
		}
	}

	// Check command's verb
	{
		have := c.Verb
		want := s.verb
		if have != want {
			t.Errorf("Invalid verb: have %q wanted %q", have, want)
		}
	}

	// Check command's nouns length and texts
	{
		have := len(c.Nouns)
		want := len(s.nouns)
		if have != want {
			t.Errorf("Nouns corrupted: have %d wanted %d", have, want)
		}
	}
	for i, want := range s.nouns {
		have := c.Nouns[i]
		if have != want {
			t.Errorf("Invalid noun: Case %d, have %q wanted %q", i, have, want)
		}
	}

	// Check command's target
	{
		have := c.Target
		want := ""
		if len(s.nouns) > 0 {
			want = s.nouns[0]
		}
		if have != want {
			t.Errorf("Invalid target: have %q wanted %q", have, want)
		}
	}

	// Check original input
	{
		have := c.Input
		want := strings.Fields(s.cmd)
		for i := range want {
			if have[i] != want[i] {
				t.Errorf("Invalid original input: word #%d, have %q wanted %q", i, have[i], want[i])
			}
		}
	}
}

func TestFuncNew(t *testing.T) {
	m := newMock()

	for _, s := range testSubjects {
		checkCommandStruct(t, m, s, New(m, s.cmd))
	}
}

func TestMethodNew(t *testing.T) {
	m := newMock()
	c := New(m, "")

	for _, s := range testSubjects {
		c.New(s.cmd)
		checkCommandStruct(t, m, s, c)
	}
}

var testMessages = [][]string{
	{"Single message test"},
	{"Hello World!", "How are you?"},
	{""},
	{"", ""},
	{"This is", "another multi-line", "test - but now", "with extra added", "lines and vitamin caffine ;)"},
}

// This tests Respond and Flush at the same time
func TestRespondAndFlush(t *testing.T) {
	for i, messages := range testMessages {
		m := newMock()
		c := New(m, "")

		for _, msg := range messages {
			c.Respond(msg)
		}
		c.Flush()

		have := m.ResponseBuf
		want := strings.Join(messages, "\n")
		if have != want {
			t.Errorf("Corrupt response: Case %d, have %q wanted %q", i, have, want)
		}
	}
}

// This tests Broadcast and Flush at the same time
func TestBroadcastAndFlush(t *testing.T) {
	for i, messages := range testMessages {
		m := newMock()
		c := New(m, "")

		for _, msg := range messages {
			c.Broadcast(nil, msg)
		}
		c.Flush()

		have := m.BroadcastBuf
		want := strings.Join(messages, "\n")
		if have != want {
			t.Errorf("Corrupt broadcast: Case %d, have %q wanted %q", i, have, want)
		}
	}
}

func TestBroadcastOmit(t *testing.T) {
	for i, messages := range testMessages {
		m := newMock()
		c := New(m, "")

		for _, msg := range messages {
			c.Broadcast([]thing.Interface{m}, msg)
		}
		c.Flush()

		have := m.BroadcastBuf
		want := ""
		if have != want {
			t.Errorf("Corrupt broadcast: Case %d, have %q wanted %q", i, have, want)
		}
	}
}

// Simplest implementation of uid.UIDLocker
type locker struct {
	uid.UID
}

func (locker) Lock()   {}
func (locker) Unlock() {}

// The main locking functions are: AddLock,	LocksModified and CanLock which are
// difficult to test on their own. So TestLocking tests all of them together.
func TestLocking(t *testing.T) {

	// Setup array to range over. Cleaner to range rather than count in for loops
	tries := [1]bool{}

	// Create a set of test lockers
	lockers := [10]locker{}
	for x := range lockers {
		lockers[x] = locker{<-uid.Next}
	}

	// Try tests twice. 1st time with things slice as created. 2nd time with
	// things slice reversed. This tests the AddLock ordering.
	for _ = range tries {

		// Create a command to manipulate the lockers with
		cmd := New(newMock(), "")

		// Range over lockers adding one at a time and checking results
		for i, locker := range lockers {

			// Check we have right number of locks
			{
				have := len(cmd.Locks)
				want := i
				if have != want {
					t.Errorf("Locks corrupted: Case %d, have %d wanted %d", i, have, want)
				}
			}

			// Check twice as LocksModified() resets when called
			for try := range tries {
				have := cmd.LocksModified()
				want := false
				if have != want {
					t.Errorf("Locks modified before add: Try %d, Case %d, have %t wanted %t", try, i, have, want)
				}
			}

			// Check what can / can't be locked before adding new lock. We should be
			// able to lock all lockers before the current one as the current one has
			// not been added yet.
			for j, l := range lockers {
				have := cmd.CanLock(l)
				want := j < i
				if have != want {
					t.Errorf("Invalid locking before add: Lock %d, Case %d, have %t wanted %t", j, i, have, want)
				}
			}

			// Add lock and check it was added
			{
				cmd.AddLock(locker)
				want := i + 1
				have := len(cmd.Locks)
				if have != want {
					t.Errorf("Lock add failed: Case %d, have %d wanted %d", i, have, want)
				}
			}

			// Check twice as LocksModified() resets when called
			for try := range tries {
				have := cmd.LocksModified()
				want := try == 0
				if have != want {
					t.Errorf("Locks modified after add: Try %d, Case %d, have %t wanted %t", try+1, i, have, want)
				}
			}

			// Check what can / can't be locked after adding new lock We should be
			// able to lock all lockers upto the current one which has just been
			// added.
			for y, h := range lockers {
				have := cmd.CanLock(h)
				want := y <= i
				if have != want {
					t.Errorf("Invalid locking after add: Lock %d, Case %d, have %t wanted %t", y, i, have, want)
				}
			}

		}

		// Reverse things slice for 2nd try - inplace without new allocations
		l := len(lockers) - 1
		for i := l / 2; i >= 0; i-- {
			lockers[i], lockers[l-i] = lockers[l-i], lockers[i]
		}
	}
}
