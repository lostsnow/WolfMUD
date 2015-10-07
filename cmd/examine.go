// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: ( EXAMINE | EXAM ) item
func Examine(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "You examine this and that, find nothing special."
		return
	}

	var (
		name = aliases[0]

		what  has.Thing
		where has.Inventory
	)

	// Work out where we are
	if a := attr.FindLocate(t); a != nil {
		where = a.Where()
	}

	// If we can, search where we are
	if where != nil {
		what = where.Search(name)
	}

	// If item not found still see if we can search narratives
	if what == nil && where != nil {
		if a := attr.FindNarrative(where.Parent()); a != nil {
			what = a.Search(name)
		}
	}

	// If item still not found try our own inventory
	if what == nil {
		if a := attr.FindInventory(t); a != nil {
			what = a.Search(name)
		}
	}

	// Was item to examine eventually found?
	if what == nil {
		msg = "You see no '" + name + "' to examine."
		return
	}

	// Check examine is not vetoed by item
	if vetoes := attr.FindVetoes(what); vetoes != nil {
		if veto := vetoes.Check("EXAMINE"); veto != nil {
			msg = veto.Message()
			return
		}
	}

	// Get item's proper name
	if n := attr.FindName(what); n != nil {
		name = n.Name()
	}

	buff := make([]byte, 0, 1024)
	buff = append(buff, "You examine "...)
	buff = append(buff, name...)
	buff = append(buff, '.')

	for _, d := range attr.FindAllDescription(what) {
		buff = append(buff, ' ')
		buff = append(buff, d.Description()...)
	}

	if i := attr.FindInventory(what); i != nil {
		buff = append(buff, ' ')
		buff = append(buff, i.List()...)
	}

	return string(buff), true
}
