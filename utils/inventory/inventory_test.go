// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package inventory

import (
	"strconv"
	"testing"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
	. "wolfmud.org/utils/test"
)

// createTestThings makes a batch of Things for testing
func createTestThings() (things []thing.Interface) {
	for x := 1; x <= 10; x++ {
		things = append(things, thing.New("Test thing "+strconv.Itoa(x), []string{"test", "thing" + strconv.Itoa(x)}, "This is a test thing."))
	}
	return
}

func TestNew(t *testing.T) {
	_ = New()
}

func TestAdd(t *testing.T) {
	things := createTestThings()
	subject := New()

	// Make sure Things added ok
	for _, thing := range things {
		l := len(subject.contents)
		subject.Add(thing)
		Equal(t, "Add", l+1, len(subject.contents))
	}

	// Check all Things are in the inventory
FOUND_ITEM:
	for _, thing := range things {
		for _, item := range subject.contents {
			if thing == item {
				continue FOUND_ITEM
			}
		}
		t.Errorf("Add '%s' not found in inventory", thing.Name())
	}

	// Check all inventory items are what we added from Things
FOUND_THING:
	for _, item := range subject.contents {
		for _, thing := range things {
			if item == thing {
				continue FOUND_THING
			}
		}
		t.Errorf("Add '%s' should not be in inventory", item.Name())
	}

	// Try adding duplicate item
	l := len(subject.contents)
	subject.Add(things[0])
	Equal(t, "Add duplicate", l, len(subject.contents))
}

func TestRemove(t *testing.T) {
	things := createTestThings()
	subject := New()

	// Try removing non-existant item
	l := len(subject.contents)
	subject.Remove(things[0])
	Equal(t, "Remove non-existant", l, len(subject.contents))

	for _, thing := range things {
		subject.Add(thing)
	}
	Equal(t, "Remove length", len(things), len(subject.contents))

	for _, thing := range things {
		subject.Remove(thing)
	}

	Equal(t, "Remove length", 0, len(subject.contents))
	Equal(t, "Remove capacity", 0, cap(subject.contents))
}

func TestFind(t *testing.T) {
	things := createTestThings()
	subject := New()

	for i, thing := range things {
		if i%2 == 1 {
			subject.Add(thing)
		}
	}

	for i, thing := range things {
		if i%2 == 1 {
			Equal(t, "find", i/2, subject.find(thing))
		} else {
			Equal(t, "find", NOT_FOUND, subject.find(thing))
		}
	}
}

func TestContains(t *testing.T) {
	things := createTestThings()
	subject := New()

	for i, thing := range things {
		if i%2 == 1 {
			subject.Add(thing)
		}
	}

	for i, thing := range things {
		Equal(t, "Contains", (i%2 == 1), subject.Contains(thing))
	}
}

func TestList(t *testing.T) {
	things := createTestThings()
	subject := New()

	for _, thing := range things {
		subject.Add(thing)
	}

	// Make sure that List = Things not omitted + Things omitted
	for x := 0; x < len(things); x++ {
		l := subject.List(things[x:]...)
		Equal(t, "List", len(things[:x]), len(l))
		for i, a := range l {
			Equal(t, "List", things[i], a)
		}
	}
}

// Define two test harnesses - but only the first can process commands
type thingHarness1 struct{ *thing.Thing }
type thingHarness2 struct{ *thing.Thing }

func (*thingHarness1) Process(cmd *command.Command) (handled bool) { return true }

func TestDelegate(t *testing.T) {

	h1 := &thingHarness1{
		Thing: thing.New("Harness 1", []string{"HARNESS1"}, "This is test harness 1"),
	}

	h2 := &thingHarness2{
		Thing: thing.New("Harness 2", []string{"HARNESS2"}, "This is test harness 2"),
	}

	subject := New()
	subject.Add(h1)

	handled := subject.Delegate(command.New(h2, "TEST"))
	Equal(t, "Delegate to another", true, handled)

	handled = subject.Delegate(command.New(h1, "TEST"))
	Equal(t, "Delegate to self", false, handled)

	subject.Remove(h1)
	subject.Add(h2)

	handled = subject.Delegate(command.New(h1, "TEST"))
	Equal(t, "Delegate to another", false, handled)

	handled = subject.Delegate(command.New(h2, "TEST"))
	Equal(t, "Delegate to self", false, handled)
}
