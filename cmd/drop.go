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

	var (
		from has.Inventory
		to   has.Inventory
		what has.Thing
	)

	// Identify inventory we want to drop something from then see if we can find
	// the something
	if a := attr.Inventory().Find(t); a != nil {
		from = a
		what = from.Search(aliases[0])
	}

	if from == nil || what == nil {
		return "You have no '" + aliases[0] + "' to drop."
	}

	// Identify where thing dropping something is
	if a := attr.Locate().Find(t); a != nil {
		if w := a.Where(); w != nil {
			if i := attr.Inventory().Find(w); i != nil {
				to = i
			}
		}
	}

	name := attr.Name().Find(what).Name()

	if to == nil {
		return "You cannot drop " + name + " here."
	}

	if veto := CheckVetoes("DROP", what); veto != nil {
		return veto.Message()
	}

	if from.Remove(what) == nil {
		return "You cannot drop " + name + "."
	}

	to.Add(what)

	return "You drop " + name + "."
}
