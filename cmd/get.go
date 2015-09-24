// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
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
		where has.Inventory
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
	what = where.Search(name)

	// If item not found in inventory also check narratives where we are
	// NOTE: Setting where to nil if we find the item prevents it from being
	// taken from the narrative inventory.
	if what == nil {
		if a := attr.FindNarrative(where.Parent()); a != nil {
			what = a.Search(name)
			where = nil
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

	// Check the get is not vetoed by the item
	if vetoes := attr.FindVetoes(what); vetoes != nil {
		if veto := vetoes.Check("GET"); veto != nil {
			msg = veto.Message()
			return
		}
	}

	// If item not from where's inventory cannot get item - most likely a
	// narrative item - we do this check after the item veto check as the veto
	// could give us a better message/reson for not being able to take the item.
	if where == nil {
		msg = "You cannot get " + name + "."
		return
	}

	// Check the get is not vetoed by the parent of the item's inventory
	if vetoes := attr.FindVetoes(where.Parent()); vetoes != nil {
		if veto := vetoes.Check("GET"); veto != nil {
			msg = veto.Message()
			return
		}
	}

	// If all seems okay try and remove item from where it is
	if where.Remove(what) == nil {
		msg = "You cannot get " + name + "."
		return
	}

	// Add item to our inventory
	to.Add(what)

	msg = "You get " + name + "."
	return msg, true
}
