// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Put(t has.Thing, aliases []string) string {

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
		return "You go to put something into something..."
	case l > 1:
		tWhat, tWhere = WhatWhere(aliases[1], t)
		tName = aliases[1]
		fallthrough
	case l == 1:
		fWhat, fWhere = WhatWhere(aliases[0], t)
		fName = aliases[0]
	}

	if fWhat == nil {
		return "You see no '" + fName + "' to put into anything."
	}

	if n := attr.Name().Find(fWhat); n != nil {
		fName = n.Name()
	}

	if fWhere != t {
		return "You don't have " + fName + " to put into anything."
	}

	if len(aliases) < 2 {
		return "What did you want to put " + fName + " into?"
	}

	if tWhere == nil {
		return "You see no '" + tName + "' to put " + fName + " into."
	}

	if n := attr.Name().Find(tWhat); n != nil {
		tName = n.Name()
	}

	if fWhat == tWhat {
		return "You can't put " + fName + " inside itself!"
	}

	fInv = attr.Inventory().Find(fWhere)
	tInv = attr.Inventory().Find(tWhat)

	if tInv == nil {
		return "You cannot put " + fName + " into " + tName + "."
	}

	fInv.Remove(fWhat)
	tInv.Add(fWhat)

	return "You put " + fName + " into " + tName + "."
}
