// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: DROP item
func Drop(s *state) {

	if len(s.words) == 0 {
		s.msg.actor.WriteString("You go to drop... something?")
		return
	}

	var (
		name = s.words[0]

		what  has.Thing
		where has.Inventory
	)

	// Search ourselves for item we want to drop
	from := attr.FindInventory(s.actor)
	if from != nil {
		what = from.Search(name)
	}

	// Was item to drop found?
	if what == nil {
		s.msg.actor.WriteJoin("You have no '", name, "' to drop.")
		return
	}

	// Find out where we are - where we are going to be dropping the item
	if a := attr.FindLocate(s.actor); a != nil {
		where = a.Where()
	}

	// Are we somewhere? We need to be somewhere so that the location can receive
	// the dropped item.
	//
	// TODO: We could drop and junk item if nowhere instead of aborting?
	if where == nil {
		s.msg.actor.WriteString("You cannot drop anything here.")
		return
	}

	// Check the drop is not vetoed by the item
	if vetoes := attr.FindVetoes(what); vetoes != nil {
		if veto := vetoes.Check("DROP"); veto != nil {
			s.msg.actor.WriteString(veto.Message())
			return
		}
	}

	// Check the drop is not vetoed by the receiving inventory
	if vetoes := attr.FindVetoes(where.Parent()); vetoes != nil {
		if veto := vetoes.Check("DROP"); veto != nil {
			s.msg.actor.WriteString(veto.Message())
			return
		}
	}

	// Get item's proper name
	if n := attr.FindName(what); n != nil {
		name = n.Name()
	}

	// Try and remove item from our inventory
	if from.Remove(what) == nil {
		s.msg.actor.WriteJoin("You cannot drop ", name, ".")
		return
	}

	// Add item to inventory where we are
	where.Add(what)

	s.msg.actor.WriteJoin("You drop ", name, ".")
	s.ok = true
}
