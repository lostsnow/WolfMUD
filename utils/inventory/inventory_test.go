// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package inventory

import (
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/utils/command"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"
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

		things[x] = &thing.Thing{}
		things[x].Unmarshal(recordjar.Record{
			"name":    "Thing " + a,
			"aliases": "test thing" + a,
			":data:":  "Test thing " + a + ".",
		})

	}
	return
}

func TestAdd(t *testing.T) {
	things := createTestThings()
	inv := Inventory{}

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
	inv := Inventory{}

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
	inv := Inventory{}

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
	inv := Inventory{}

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
	inv := Inventory{}

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
	inv := Inventory{}

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

// Define two test harnesses
type willProcess struct{ *thing.Thing }
type wontProcess struct{ *thing.Thing }

// Implement command.Interface on willProcess so that only it CAN process
// commands wontProcess will NOT have a Process method and CANNOT process
// commands.
func (*willProcess) Process(cmd *command.Command) (handled bool) { return true }

func TestProcess(t *testing.T) {

	// Setup 'will' process
	will := &willProcess{&thing.Thing{}}
	will.Thing.Unmarshal(recordjar.Record{
		"name":    "Harness 1",
		"aliases": "HARNESS1",
		":data:":  "This is test harness 1.",
	})

	// Setup 'wont' process
	wont := &wontProcess{&thing.Thing{}}
	wont.Thing.Unmarshal(recordjar.Record{
		"name":    "Harness 2",
		"aliases": "HARNESS2",
		":data:":  "This is test harness 2.",
	})

	// Test with 'will' which can process commands
	inv := Inventory{}
	inv.Add(will)

	// Check recursion. 'will' should not be delegated to when also issuing command
	{
		have := inv.Process(command.New(will, "TEST"))
		want := false
		if have != want {
			t.Errorf("Process mis-handled: have %t wanted %t", have, want)
		}
	}

	// 'will' should handle command from 'wont'
	{
		have := inv.Process(command.New(wont, "TEST"))
		want := true
		if have != want {
			t.Errorf("Process not handled: have %t wanted %t", have, want)
		}
	}

	// Test with 'wont' which cannot process commands
	inv.Remove(will)
	inv.Add(wont)

	// 'wont' cannot handle command from 'will'
	{
		have := inv.Process(command.New(will, "TEST"))
		want := false
		if have != want {
			t.Errorf("Process mis-handled: have %t wanted %t", have, want)
		}
	}

	// 'wont' cannot handle command from self
	{
		have := inv.Process(command.New(wont, "TEST"))
		want := false
		if have != want {
			t.Errorf("Process mis-handled: have %t wanted %t", have, want)
		}
	}
}

func TestLength(t *testing.T) {
	things := createTestThings()
	inv := Inventory{}

	// Add things and check expected length of the inventory
	for i, thing := range things {
		inv.Add(thing)
		have := inv.Length()
		want := i + 1
		if have != want {
			t.Errorf("Invalid length: have %t wanted %t", have, want)
		}
	}
}
