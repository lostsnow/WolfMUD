/*
	Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.

	Use of this source code is governed by the license in the LICENSE file
	included with the source code.
*/

package entities

import (
	"reflect"
	"strings"
	"testing"
)

var testData = [][]string{
	{"A ball", "BALL", "This is a small ball."},
	{"A curious brass lattice", "LATTICE", "This is a finely crafted, intricate lattice of fine brass wires forming a roughly ball shaped curiosity."},
}

func TestNewThing(t *testing.T) {
	for _, row := range testData {
		t1 := NewThing(row[0], row[1], row[2])
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
}

func TestName(t *testing.T) {
	for _, row := range testData {
		t1 := NewThing(row[0], row[1], row[2])
		expected := row[0]
		if result := t1.Name(); result != expected {
			t.Errorf("Name: Wrong name, expected: %s got: %#v", expected, result)
		}
	}
}

func TestAlias(t *testing.T) {
	for _, row := range testData {
		t1 := NewThing(row[0], row[1], row[2])
		expected := strings.ToUpper(row[1])
		if result := t1.Alias(); result != expected {
			t.Errorf("Alias: Wrong alias, expected: %s got: %#v", expected, result)
		}
	}
}

func ExampleProcess_examine() {
	commandHelper("EXAMINE")
	// Output:
	// You examine A ball. This is a small ball.
	// You examine A curious brass lattice. This is a finely crafted, intricate lattice of fine brass wires forming a roughly ball shaped curiosity.
}

func ExampleProcess_ex() {
	commandHelper("EX")
	// Output:
	// You examine A ball. This is a small ball.
	// You examine A curious brass lattice. This is a finely crafted, intricate lattice of fine brass wires forming a roughly ball shaped curiosity.
}

func commandHelper(cmd string) (handled bool) {
	for _, row := range testData {
		t1 := NewThing(row[0], row[1], row[2])
		handled = t1.Process(NewCommand(t1, cmd+" "+t1.Alias()))
	}
	return
}
