// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"strings"
)

// Syntax: GET item
func Get(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "You go to get... something?"
		return
	}

	var (
		what  has.Thing // The item we want to get
		where has.Thing // Where the item currently is
	)

	name := aliases[0]

	// Work out where we are and then search for item to get there
	if a := attr.Locate().Find(t); a != nil {
		where = a.Where()
		what = search(name, where)
	}

	if what == nil {
		msg = "You see no '" + name + "' to get."
		return
	}

	// Check we have an inventory and can carry things
	to := attr.Inventory().Find(t)
	if to == nil {
		msg = "You can't carry anything!"
		return
	}

	// Check the get is not vetoed by the item
	if veto := CheckVetoes("GET", what); veto != nil {
		msg = veto.Message()
		return
	}

	// Get item's proper name
	if n := attr.Name().Find(what); n != nil {
		name = n.Name()
	}

	// NOTE: If we try to get a narrative item it won't be found in the
	// inventory, it's in the narrative, so the remove on the inventory will
	// fail.
	from := attr.Inventory().Find(where)
	if from.Remove(what) == nil {
		msg = strings.Title(name[0:1]) + name[1:] + " cannot be taken."
		return
	}

	// Add item to our inventory
	to.Add(what)

	msg = "You get " + name + "."
	return msg, true
}
