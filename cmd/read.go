// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: READ item
func init() {
	AddHandler(Read, "READ")
}

func Read(s *state) {

	if len(s.words) == 0 {
		s.msg.actor.WriteString("Did you want to read something specific?")
		return
	}

	var (
		name = s.words[0]

		what    has.Thing
		where   has.Inventory
		writing string
	)

	// Work out where we are
	if a := attr.FindLocate(s.actor); a != nil {
		where = a.Where()
	}

	// If we are somewhere and item not found yet try searching inventory where
	// we are
	if where != nil && what == nil {
		what = where.Search(name)
	}

	// If we are somewhere and item still not found try searching narratives
	// where we are
	if where != nil && what == nil {
		if a := attr.FindNarrative(where.Parent()); a != nil {
			what = a.Search(name)
		}
	}

	// If item still not found try our own inventory
	if what == nil {
		if a := attr.FindInventory(s.actor); a != nil {
			what = a.Search(name)
		}
	}

	// Was item to read found?
	if what == nil {
		s.msg.actor.WriteJoin("You see no '", name, "' to read.")
		return
	}

	// Get item's proper name
	if n := attr.FindName(what); n != nil {
		name = n.Name()
	}

	// Find if item has writing
	if a := attr.FindWriting(what); a != nil {
		writing = a.Writing()
	}

	// Was writing found?
	if writing == "" {
		s.msg.actor.WriteJoin("You see no writing on ", name, " to read.")
		return
	}

	s.msg.actor.WriteJoin("You read the writing on ", name, ". It says: ", writing)
	s.ok = true
}
