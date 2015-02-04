// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Take(t has.Thing, aliases []string) string {

	var (
		fWhat has.Thing
		fName string
		fInv  has.Inventory

		tWhat has.Thing
		tName string
		tInv  has.Inventory
	)

	switch l := len(aliases); {
	case l == 0:
		return "You go to take something out of something else..."
	case l > 1:
		tWhat, _ = WhatWhere(aliases[1], t)
		tName = aliases[1]
		fallthrough
	case l == 1:
		fName = aliases[0]
	}

	if tWhat == nil {
		return "You don't see '" + tName + "' to take things out of."
	}

	if n := attr.Name().Find(tWhat); n != nil {
		tName = n.Name()
	}

	tInv = attr.Inventory().Find(tWhat)

	if tInv == nil {
		return tName + " does not contain anything."
	}

	fWhat = tInv.Search(aliases[0])

	if fWhat == nil {
		return "You see no '" + fName + "' in " + tName + "."
	}

	if n := attr.Name().Find(fWhat); n != nil {
		fName = n.Name()
	}

	fInv = attr.Inventory().Find(t)

	tInv.Remove(fWhat)
	fInv.Add(fWhat)

	return "You take " + fName + " out of " + tName + "."
}
