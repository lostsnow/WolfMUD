// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: DROP item
func init() {
	AddHandler(Drop, "DROP")
}

func Drop(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.Send("You go to drop... something?")
		return
	}

	name := s.words[0]

	// Search ourselves for item we want to drop
	from := attr.FindInventory(s.actor)

	// Are we carrying anything at all?
	if from.Empty() {
		s.msg.Actor.Send("You don't have anything to drop.")
		return
	}

	what := from.Search(name)

	// Was item to drop found?
	if what == nil {
		s.msg.Actor.Send("You have no '", name, "' to drop.")
		return
	}

	// Are we somewhere? We need to be somewhere so that the location can receive
	// the dropped item.
	//
	// TODO: We could drop and junk item if nowhere instead of aborting?
	if s.where == nil {
		s.msg.Actor.Send("You cannot drop anything here.")
		return
	}

	// Check the drop is not vetoed by the item
	if veto := attr.FindVetoes(what).Check("DROP"); veto != nil {
		s.msg.Actor.Send(veto.Message())
		return
	}

	// Check the drop is not vetoed by the receiving inventory
	if veto := attr.FindVetoes(s.where.Parent()).Check("DROP"); veto != nil {
		s.msg.Actor.Send(veto.Message())
		return
	}

	// Get item's proper name
	name = attr.FindName(what).Name(name)

	// Try and remove item from our inventory
	if from.Remove(what) == nil {
		s.msg.Actor.Send("You cannot drop ", name, ".")
		return
	}

	// Add item to inventory where we are
	s.where.Add(what)

	who := attr.FindName(s.actor).Name("Someone")

	s.msg.Actor.Send("You drop ", name, ".")
	s.msg.Observer.Send(who, " drops ", name, ".")
	s.ok = true
}
