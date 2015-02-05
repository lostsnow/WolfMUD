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

	var (
		tWhat  has.Thing // Info for thing we are putting into something
		tWhere has.Thing
		tName  string

		cWhat  has.Thing // Info for container we want to put something into
		cWhere has.Thing
		cName  string
		cInv   has.Inventory
	)

	switch l := len(aliases); {
	case l == 0:
		msg = "You go to put something into something..."
		return
	case l > 1:
		// Try and identify container
		cWhat, cWhere = WhatWhere(aliases[1], t)
		cName = aliases[1]
		fallthrough
	case l == 1:
		// Try and identify item
		tWhat, tWhere = WhatWhere(aliases[0], t)
		tName = aliases[0]
	}

	if tWhat == nil {
		msg = "You see no '" + tName + "' to put into anything."
		return
	}

	if n := attr.Name().Find(tWhat); n != nil {
		tName = n.Name()
	}

	if tWhere != t {
		msg = "You don't have " + tName + " to put into anything."
		return
	}

	if len(aliases) < 2 {
		msg = "What did you want to put " + tName + " into?"
		return
	}

	if cWhere == nil {
		msg = "You see no '" + cName + "' to put " + tName + " into."
		return
	}

	if n := attr.Name().Find(cWhat); n != nil {
		cName = n.Name()
	}

	if tWhat == cWhat {
		msg = "You can't put " + tName + " inside itself!"
		return
	}

	if cInv = attr.Inventory().Find(cWhat); cInv == nil {
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

	attr.Inventory().Find(tWhere).Remove(tWhat)
	cInv.Add(tWhat)

	msg = "You put " + tName + " into " + cName + "."
	return msg, true
}
