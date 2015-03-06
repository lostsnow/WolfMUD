// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Syntax: READ item
func Read(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "Did you want to read something specific?"
		return
	}

	var (
		name = aliases[0]

		what    has.Thing
		where   has.Thing
		writing string
	)

	// Work out where we are
	if a := attr.FindLocate(t); a != nil {
		where = a.Where()
	}

	// Are we somewhere?
	if where != nil {
		// Search for item in inventory where we are
		if a := attr.FindInventory(where); a != nil {
			what = a.Search(name)
		}

		// If item not found in inventory try searching narratives
		if what == nil {
			if a := attr.FindNarrative(where); a != nil {
				what = a.Search(name)
			}
		}
	}

	// If item still not found try our own inventory
	if what == nil {
		if a := attr.FindInventory(t); a != nil {
			what = a.Search(name)
		}
	}

	// Was item to read found?
	if what == nil {
		msg = "You see no '" + name + "' to read."
		return
	}

	// Get item's proper name
	if n := attr.FindName(what); n != nil {
		name = n.Name()
	}

	// Find if item has writing
	if a := attr.FindWriting(what); a != nil {
		writing = a.Writing()
	}

	// Was writing found?
	if writing == "" {
		msg = "You see no writing on " + name + " to read."
		return
	}

	msg = "You read the writing on " + name + ". It says: " + writing
	return msg, true
}
