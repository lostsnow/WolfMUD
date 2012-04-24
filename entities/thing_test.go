/*
	Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.

	Use of this source code is governed by the license in the LICENSE file
	included with the source code.
*/

package entities

import (
	"reflect"
	"testing"
	"runtime"
)

func TestNewThing(t *testing.T) {

	t1 := NewThing("A ball", "BALL", "This is a test ball.")

	expected := "*entities.thing"
	if reflect.TypeOf(t1).String() != expected {
		t.Errorf("NewThing: Wrong type, expected: %s got %T: %#v", expected, t1, t1)
	}
	if _, ok := t1.(Examiner); !ok {
		t.Errorf("NewThing: Does not implement Examiner interface")
	}
	if _, ok := t1.(Processor); !ok {
		t.Errorf("NewThing: Does not implement Processor interface")
	}
}

func TestName(t *testing.T) {

	t1 := NewThing("A ball", "BALL", "This is a test ball.")

	expected := "A ball"
	if result := t1.Name(); result != expected {
		t.Errorf("Name: Wrong name, expected: %s got: %#v", expected, result)
	}
}

func TestAlias(t *testing.T) {
	t1 := NewThing("A ball", "BALL", "This is a test ball.")

	expected := "BALL"
	if result := t1.Alias(); result != expected {
		t.Errorf("Alias: Wrong alias, expected: %s got: %#v", expected, result)
	}
}

func TestProcess_examine(t *testing.T) {
	l1 := NewLocation("Test Room", "TEST1", "This is a test room")
	p1 := NewPlayer("Bob", "Bob", "This is Bob.")
	t1 := NewThing("A ball", "BALL", "This is a test ball.")
	l1.Add(t1)
	l1.Add(p1)
	p1.Locate(l1)
	t1.Locate(l1)

	handled := false
	response := ""
	expected := ""

	// Bypass Input() and Process directly
	handled = p1.Process(NewCommand(p1, "EXAMINE BALL"))
	runtime.Gosched()
	response = p1.Output()
	expected = "You examine A ball. This is a test ball."

	if response != expected {
		t.Errorf("EXAMINE: wrong response\nGOT: |%s|\nEXP: |%s|\n", response, expected)
	}
	if !handled {
		t.Errorf("EXAMINE: not handled: EXAMINE BALL\n")
	}

	// Bypass Input() and Process directly
	handled = p1.Process(NewCommand(p1, "EXAMINE BANANA"))
	runtime.Gosched()
	response = p1.Output()
	expected = ""

	if response != expected {
		t.Errorf("EXAMINE: wrong response\nGOT: |%s|\nEXP: |%s|\n", response, expected)
	}
	if handled {
		t.Errorf("EXAMINE: handled: EXAMINE BANANA\n")
	}
}
