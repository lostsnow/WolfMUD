// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Syntax: GET item
func Get(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "You go to get... something?"
		return
	}

	var (
		name = aliases[0]

		what  has.Thing
		where has.Thing
	)

	// Work out where we are
	if a := attr.FindLocate(t); a != nil {
		where = a.Where()
	}

	// Are we somewhere?
	// NOTE: The only way this should be possible is if something is gotten when
	// the current thing is not in the world - but then where would it come from?
	if where == nil {
		msg = "You cannot get anything here."
		return
	}

	// Search for item we want to get in the inventory where we are
	from := attr.FindInventory(where)
	if from != nil {
		what = from.Search(name)
		if what == nil {
			from = nil
		}
	}

	// If item not found in inventory also check narratives where we are
	if what == nil {
		if a := attr.FindNarrative(where); a != nil {
			what = a.Search(name)
		}
	}

	// Was item to get found?
	if what == nil {
		msg = "You see no '" + name + "' to get."
		return
	}

	// Check we have an inventory so we can carry things
	to := attr.FindInventory(t)
	if to == nil {
		msg = "You can't carry anything!"
		return
	}

	// Get item's proper name
	if a := attr.FindName(what); a != nil {
		name = a.Name()
	}

	// Check item not trying to get itself
	if what == t {
		msg = "Trying to pick youreself up by your bootlaces?"
		return
	}

	// If item not from where's inventory cannot get item - most likely a
	// narrative item
	if from == nil {
		msg = "You cannot get " + name + "."
		return
	}

	// Check the get is not vetoed by the item
	if vetoes := attr.FindVetoes(what); vetoes != nil {
		if veto := vetoes.Check("GET"); veto != nil {
			msg = veto.Message()
			return
		}
	}

	// Check the get is not vetoed by the parent of the item's inventory
	if vetoes := attr.FindVetoes(where); vetoes != nil {
		if veto := vetoes.Check("GET"); veto != nil {
			msg = veto.Message()
			return
		}
	}

	// If all seems okay try and remove item from where it is
	if from.Remove(what) == nil {
		msg = "You cannot get " + name + "."
		return
	}

	// Add item to our inventory
	to.Add(what)

	msg = "You get " + name + "."
	return msg, true
}
