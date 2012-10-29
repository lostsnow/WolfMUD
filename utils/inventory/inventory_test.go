// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package inventory

import (
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/utils/command"
	"strconv"
	"testing"
)

const (
	MAX_TEST_THINGS = 10
)

// createTestThings makes a batch of Things for testing
func createTestThings() (things []thing.Interface) {

	things = make([]thing.Interface, MAX_TEST_THINGS, MAX_TEST_THINGS)

	for x := 0; x < MAX_TEST_THINGS; x++ {

		a := strconv.Itoa(x)

		things[x] = thing.New(
			"Thing "+a, []string{"test", "thing" + a}, "Test thing "+a+".",
		)

	}
	return
}

func TestNew(t *testing.T) {
	inv := New()

	if inv == nil {
		t.Errorf("New inventory not created!")
	}

	if len(inv.contents) != 0 {
		t.Errorf("New inventory not empty!")
	}
}

func TestAdd(t *testing.T) {
	things := createTestThings()
	inv := New()

	// Make sure Things added ok
	for i, thing := range things {
		inv.Add(thing)
		have := len(inv.contents)
		want := i + 1
		if have != want {
			t.Errorf("Invalid inventory size: Case %d, have %d wanted %d", i, have, want)
		}
	}

	// Check inventory only contains test subjects
FOUND_ITEM:
	for i, have := range things {
		for _, want := range inv.contents {
			if have == want {
				continue FOUND_ITEM
			}
		}
		t.Errorf("Invalid item: Case %d, have %#v", i, have)
	}

	// Check all test subjects are in the inventory
FOUND_THING:
	for i, have := range inv.contents {
		for _, want := range things {
			if have == want {
				continue FOUND_THING
			}
		}
		t.Errorf("Missing item: Case %d, have %#v", i, have)
	}

	// Try adding duplicate item
	want := len(inv.contents)
	inv.Add(things[0])
	have := len(inv.contents)
	if have != want {
		t.Errorf("Duplicate item added: have %d want %d", have, want)
	}
}

func TestRemoveNotExist(t *testing.T) {
	things := createTestThings()
	inv := New()

	inv.Add(things[0])
	want := len(inv.contents)
	inv.Remove(things[1])
	have := len(inv.contents)
	if have != want {
		t.Errorf("Removed non-existant: have %d wanted %d", have, want)
	}
}

// When inventory emptied length and capacity should be zero
func TestRemoveEmpty(t *testing.T) {
	things := createTestThings()
	inv := New()

	// Add all things
	for _, thing := range things {
		inv.Add(thing)
	}

	// Remove all things
	for _, thing := range things {
		inv.Remove(thing)
	}

	// Check length
	{
		have := len(inv.contents)
		want := 0
		if have != want {
			t.Errorf("Wrong length: have %d wanted %d", have, want)
		}
	}

	// Check capacity
	{
		have := cap(inv.contents)
		want := 0
		if have != want {
			t.Errorf("Wrong capacity: have %d wanted %d", have, want)
		}
	}
}

func TestFind(t *testing.T) {
	things := createTestThings()
	inv := New()

	// Add odd things
	for i, thing := range things {
		if i%2 == 1 {
			inv.Add(thing)
		}
	}

	// Check we can find all odd things we added
	for i, thing := range things {
		have := inv.find(thing)
		want := NOT_FOUND
		if i%2 == 1 {
			want = i / 2
		}
		if have != want {
			t.Errorf("Invalid find: Case %d, have %d wanted %d", i, have, want)
		}
	}
}

func TestContains(t *testing.T) {
	things := createTestThings()
	inv := New()

	// Add odd things
	for i, thing := range things {
		if i%2 == 1 {
			inv.Add(thing)
		}
	}

	// Check only odd things found
	for i, thing := range things {
		have := inv.Contains(thing)
		want := i%2 == 1
		if have != want {
			t.Errorf("Invalid contains: Case %d, have %t wanted %t", i, have, want)
		}
	}
}

func TestList(t *testing.T) {
	things := createTestThings()
	inv := New()

	for _, thing := range things {
		inv.Add(thing)
	}

	// Make sure that List contains Things not omitted
	for i := 0; i < len(things); i++ {
		omitted := things[i:]
		included := things[:i]
		list := inv.List(omitted...)

		//t.Errorf("\n%#v\n%#v\n\n", list, included)
		//continue

		{
			want := len(things) - len(omitted)
			have := len(list)
			if have != want {
				t.Errorf("List length corrupted: Case %d, have %d wanted %d", i, have, want)
			}
		}

		// Make sure all included things in list
	FOUND_IN_LIST:
		for i, want := range included {
			for _, have := range list {
				if have == want {
					continue FOUND_IN_LIST
				}
			}
			t.Errorf("List missing item: Case %d, wanted %#v", i, want)
		}

		// Make sure list contains only included things
	FOUND_IN_INCLUDED:
		for i, want := range list {
			for _, have := range included {
				if have == want {
					continue FOUND_IN_INCLUDED
				}
			}
			t.Errorf("Invalid item in list: Case %d, wanted %#v", i, want)
		}

	}
}

// Define test harness that CAN process commands
type thingHarness1 struct{ *thing.Thing }

func (*thingHarness1) Process(cmd *command.Command) (handled bool) { return true }

// Define test harness that CANNOT process commands
type thingHarness2 struct{ *thing.Thing }

func TestDelegate(t *testing.T) {

	h1 := &thingHarness1{
		Thing: thing.New("Harness 1", []string{"HARNESS1"}, "This is test harness 1"),
	}

	h2 := &thingHarness2{
		Thing: thing.New("Harness 2", []string{"HARNESS2"}, "This is test harness 2"),
	}

	// Test with h1 which can process commands
	inv := New()
	inv.Add(h1)

	// Check recursion. h1 should not be delegated to when also issuing command
	{
		have := inv.Delegate(command.New(h1, "TEST"))
		want := false
		if have != want {
			t.Errorf("Delegation mis-handled: have %t wanted %t", have, want)
		}
	}

	// h1 should handle command from h2
	{
		have := inv.Delegate(command.New(h2, "TEST"))
		want := true
		if have != want {
			t.Errorf("Delegation not handled: have %t wanted %t", have, want)
		}
	}

	// Test with h2 which cannot process commands
	inv.Remove(h1)
	inv.Add(h2)

	// h2 cannot handle command from h1
	{
		have := inv.Delegate(command.New(h1, "TEST"))
		want := false
		if have != want {
			t.Errorf("Delegation mis-handled: have %t wanted %t", have, want)
		}
	}

	// h2 cannot handle command from self
	{
		have := inv.Delegate(command.New(h2, "TEST"))
		want := false
		if have != want {
			t.Errorf("Delegation mis-handled: have %t wanted %t", have, want)
		}
	}
}
