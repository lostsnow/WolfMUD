// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Put(t has.Thing, aliases []string) (msg string, ok bool) {

	var (
		fWhat  has.Thing
		fWhere has.Thing
		fName  string
		fInv   has.Inventory

		tWhat  has.Thing
		tWhere has.Thing
		tName  string
		tInv   has.Inventory
	)

	switch l := len(aliases); {
	case l == 0:
		msg = "You go to put something into something..."
		return
	case l > 1:
		tWhat, tWhere = WhatWhere(aliases[1], t)
		tName = aliases[1]
		fallthrough
	case l == 1:
		fWhat, fWhere = WhatWhere(aliases[0], t)
		fName = aliases[0]
	}

	if fWhat == nil {
		msg = "You see no '" + fName + "' to put into anything."
		return
	}

	if n := attr.Name().Find(fWhat); n != nil {
		fName = n.Name()
	}

	if fWhere != t {
		msg = "You don't have " + fName + " to put into anything."
		return
	}

	if len(aliases) < 2 {
		msg = "What did you want to put " + fName + " into?"
		return
	}

	if tWhere == nil {
		msg = "You see no '" + tName + "' to put " + fName + " into."
		return
	}

	if n := attr.Name().Find(tWhat); n != nil {
		tName = n.Name()
	}

	if fWhat == tWhat {
		msg = "You can't put " + fName + " inside itself!"
		return
	}

	fInv = attr.Inventory().Find(fWhere)
	tInv = attr.Inventory().Find(tWhat)

	if tInv == nil {
		msg = "You cannot put " + fName + " into " + tName + "."
		return
	}

	// Check for veto on item being put into container
	if veto := CheckVetoes("PUT", fWhat); veto != nil {
		msg = veto.Message()
		return
	}

	// Check for veto on container
	if veto := CheckVetoes("PUT", tWhat); veto != nil {
		msg = veto.Message()
		return
	}

	fInv.Remove(fWhat)
	tInv.Add(fWhat)

	msg = "You put " + fName + " into " + tName + "."
	return msg, true
}
