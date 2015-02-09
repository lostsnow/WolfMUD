// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Syntax: PUT item container
func Put(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "You go to put something into something else..."
		return
	}

	tName := aliases[0]

	// Search ourselves for item to put into container
	tWhat := search(tName, t)

	if tWhat == nil {
		msg = "You have no '" + tName + "' to put into anything."
		return
	}

	// Get item's proper name
	if n := attr.Name().Find(tWhat); n != nil {
		tName = n.Name()
	}

	// Check a container was specified
	if len(aliases) < 2 {
		msg = "What did you want to put " + tName + " into?"
		return
	}

	// Try and find container
	var (
		cName = aliases[1]
		cWhat = what(cName, t)
	)

	// Was container found?
	if cWhat == nil {
		msg = "You see no '" + cName + "' to put " + tName + " into."
		return
	}

	// Unless our name is Klein we can't put something inside itself! ;)
	if tWhat == cWhat {
		msg = "You can't put " + tName + " inside itself!"
		return
	}

	// Get container's proper name
	if n := attr.Name().Find(cWhat); n != nil {
		cName = n.Name()
	}

	// Check container is actually a container with an inventory
	cInv := attr.Inventory().Find(cWhat)
	if cInv == nil {
		msg = "You cannot put " + tName + " into " + cName + "."
		return
	}

	// Check for veto on item being put into container
	if veto := CheckVetoes("PUT", tWhat); veto != nil {
		msg = veto.Message()
		return
	}

	// Make sure nothing would stop us letting go of item
	if veto := CheckVetoes("DROP", tWhat); veto != nil {
		msg = veto.Message()
		return
	}

	// Check for veto on container
	if veto := CheckVetoes("PUT", cWhat); veto != nil {
		msg = veto.Message()
		return
	}

	// Remove item from where it is
	if a := attr.Inventory().Find(t); a != nil {
		if a.Remove(tWhat) == nil {
			msg = "Something stops you putting " + tName + " anywhere."
			return
		}
	}

	// Put item into comtainer
	cInv.Add(tWhat)

	msg = "You put " + tName + " into " + cName + "."
	return msg, true
}
