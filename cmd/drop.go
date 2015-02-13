// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Syntax: DROP item
func Drop(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "You go to drop... something?"
		return
	}

	var (
		name = aliases[0]

		what  has.Thing
		where has.Thing
	)

	// Search ourselves for item we want to drop
	from := attr.Inventory().Find(t)
	if from != nil {
		what = from.Search(name)
	}

	// Was item to drop found?
	if what == nil {
		msg = "You have no '" + name + "' to drop."
		return
	}

	// Find out where we are - where we are going to be dropping the item
	if a := attr.Locate().Find(t); a != nil {
		where = a.Where()
	}

	// Are we somewhere?
	// TODO: We could drop and junk item if nowhere instead of aborting?
	if where == nil {
		msg = "You cannot drop anything here."
		return
	}

	// Check inventory available to receive dropped item
	// NOTE: The only way this should be possible is if something is dropped when
	// the current thing is not in the world.
	to := attr.Inventory().Find(where)
	if to == nil {
		msg = "You cannot drop anything here."
		return
	}

	// Check the drop is not vetoed by the item
	if veto := CheckVetoes("DROP", what); veto != nil {
		msg = veto.Message()
		return
	}

	// Get item's proper name
	if n := attr.Name().Find(what); n != nil {
		name = n.Name()
	}

	// Try and remove item from our inventory
	if from.Remove(what) == nil {
		msg = "You cannot drop " + name + "."
		return
	}

	// Add item to inventory where we are
	to.Add(what)

	msg = "You drop " + name + "."
	return msg, true
}
