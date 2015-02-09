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

	name := aliases[0]

	// Search ourselves for item we want to drop
	what := search(name, t)

	if what == nil {
		msg = "You have no '" + name + "' to drop."
		return
	}

	// Get item's proper name
	if n := attr.Name().Find(what); n != nil {
		name = n.Name()
	}

	// Find our own inventory we are dropping item from
	from := attr.Inventory().Find(t)

	// Find out where we are - where we are going to be dropping the item
	var to has.Inventory
	if a := attr.Locate().Find(t); a != nil {
		if w := a.Where(); w != nil {
			if i := attr.Inventory().Find(w); i != nil {
				to = i
			}
		}
	}

	if to == nil {
		msg = "You cannot drop " + name + " here."
		return
	}

	// Check the drop is not vetoed by the item
	if veto := CheckVetoes("DROP", what); veto != nil {
		msg = veto.Message()
		return
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
