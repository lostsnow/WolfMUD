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
func init() {
	AddHandler(Examine, "EXAM", "EXAMINE")
}

func Examine(s *state) {

	if len(s.words) == 0 {
		s.msg.actor.WriteString("You examine this and that, find nothing special.")
		return
	}

	var (
		name = s.words[0]

		what  has.Thing
		where has.Inventory
	)

	// Work out where we are
	if a := attr.FindLocate(s.actor); a != nil {
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
		if a := attr.FindInventory(s.actor); a != nil {
			what = a.Search(name)
		}
	}

	// Was item to examine eventually found?
	if what == nil {
		s.msg.actor.WriteJoin("You see no '", name, "' to examine.")
		return
	}

	// Check examine is not vetoed by item
	if vetoes := attr.FindVetoes(what); vetoes != nil {
		if veto := vetoes.Check("EXAMINE"); veto != nil {
			s.msg.actor.WriteString(veto.Message())
			return
		}
	}

	// Get item's proper name
	if n := attr.FindName(what); n != nil {
		name = n.Name()
	}

	s.msg.actor.WriteJoin("You examine ", name, ".")

	for _, d := range attr.FindAllDescription(what) {
		s.msg.actor.WriteJoin(" ", d.Description())
	}

	if i := attr.FindInventory(what); i != nil {
		s.msg.actor.WriteJoin(" ", i.List())
	}

	s.ok = true
}
