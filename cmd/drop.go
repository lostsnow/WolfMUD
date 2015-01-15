// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Drop(t has.Thing, aliases []string) string {

	if len(aliases) == 0 {
		return "You go to drop... something?"
	}

	if t == nil {
		return "You have no '" + aliases[0] + "' to drop."
	}

	var from has.Inventory
	var to has.Inventory
	var what has.Thing

	// Identify inventory we want to drop something from then see if we can find
	// the something
	if a := attr.FindInventory(t); a != nil {
		from = a
		what = from.Find(aliases[0])
	}

	if from == nil || what == nil {
		return "You have no '" + aliases[0] + "' to drop."
	}

	// Identify location of thing dropping something
	if a := attr.FindLocate(t); a != nil {
		if l := a.Location(); l != nil {
			if i := attr.FindInventory(l); i != nil {
				to = i
			}
		}
	}

	name := attr.FindName(what).Name()

	if to == nil {
		return "You cannot drop " + name + " here."
	}

	if msg, vetoed := CheckVetoes("DROP", what); vetoed {
		return msg
	}

	if from.Remove(what) == nil {
		return "You cannot drop " + name + "."
	}

	to.Add(what)

	return "You drop " + name + "."
}
