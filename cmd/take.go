// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Take(t has.Thing, aliases []string) (msg string, ok bool) {

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
		msg = "You go to take something out of something else..."
		return
	case l > 1:
		tWhat, _ = WhatWhere(aliases[1], t)
		tName = aliases[1]
		fallthrough
	case l == 1:
		fName = aliases[0]
	}

	if tWhat == nil {
		msg = "You don't see '" + tName + "' to take things out of."
		return
	}

	if n := attr.Name().Find(tWhat); n != nil {
		tName = n.Name()
	}

	tInv = attr.Inventory().Find(tWhat)

	if tInv == nil {
		msg = tName + " does not contain anything."
		return
	}

	fWhat = tInv.Search(aliases[0])

	if fWhat == nil {
		msg = "You see no '" + fName + "' in " + tName + "."
		return
	}

	if n := attr.Name().Find(fWhat); n != nil {
		fName = n.Name()
	}

	fInv = attr.Inventory().Find(t)

	tInv.Remove(fWhat)
	fInv.Add(fWhat)

	msg = "You take " + fName + " out of " + tName + "."
	return msg, true
}
