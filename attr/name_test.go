// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr_test

import (
	"testing"

	. "code.wolfmud.org/WolfMUD.git/attr"
)

func TestTheName(t *testing.T) {
	for _, test := range []struct {
		data string
		want string
	}{
		{"", "Something"},
		{"a", "a"},
		{"A", "A"},
		{"an", "an"},
		{"An", "An"},
		{"AN", "AN"},
		{"a ", "the "},
		{"A ", "The "},
		{"an ", "the "},
		{"An ", "The "},
		{"AN ", "THE "},
		{"a frog", "the frog"},
		{"A frog", "The frog"},
		{"an apple", "the apple"},
		{"An apple", "The apple"},
		{"AN APPLE", "THE APPLE"},
		{"some apples", "the apples"},
		{"Some apples", "The apples"},
		{"SOME APPLES", "THE APPLES"},
		{"apples", "apples"},
	} {
		t.Run(test.data, func(t *testing.T) {
			n := NewName(test.data)
			have := n.TheName("Something")

			if have != test.want {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

// Test case for a nil *Name with a preset
func TestTheName_nil(t *testing.T) {
	have := (*Name)(nil).TheName("Something")
	if have != "Something" {
		t.Errorf("\nhave %+q\nwant \"Something\"", have)
	}
}

// Test case for a nil *Name and an empty string for the preset
func TestTheName_nil_empty(t *testing.T) {
	have := (*Name)(nil).TheName("")
	if have != "" {
		t.Errorf("\nhave %+q\nwant \"\"", have)
	}
}

func BenchmarkString(b *testing.B) {
	for _, test := range []string{

		// NOTE: Keep the test strings all the same length so that the
		// results can be compared with each other - i.e. changes, non-changes and
		// quick paths.

		// Matches, changes
		"a rabbit",
		"an apple",
		"some ink",

		// Only first letter matches, no changes (slow path)
		"aardvark",
		"sometime",

		// Non 1st letter match, no changes (fast path)
		"nine oak",
		"dead elk",

		// Only space detected at correct position, no changes (fast path)
		"I robots",
		"to these",
		"four bee",
	} {
		n := NewName(test)
		b.Run(test, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = n.TheName("Something")
			}
		})
	}
}
