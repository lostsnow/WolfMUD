// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Drop(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "You go to drop... something?"
		return
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
		msg = "You have no '" + aliases[0] + "' to drop."
		return
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
		msg = "You cannot drop " + name + " here."
		return
	}

	if veto := CheckVetoes("DROP", what); veto != nil {
		msg = veto.Message()
		return
	}

	if from.Remove(what) == nil {
		msg = "You cannot drop " + name + "."
		return
	}

	to.Add(what)

	msg = "You drop " + name + "."
	return msg, true
}
