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
	addHandler(drop{}, "DROP")
}

type drop cmd

func (drop) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to drop... something?")
		return
	}

	name := s.words[0]

	// Search ourselves for item we want to drop
	from := attr.FindInventory(s.actor)

	// Are we carrying anything at all?
	if from.Empty() {
		s.msg.Actor.SendBad("You don't have anything to drop.")
		return
	}

	what := from.Search(name)

	// Was item to drop found?
	if what == nil {
		s.msg.Actor.SendBad("You have no '", name, "' to drop.")
		return
	}

	// Check the drop is not vetoed by the item
	if veto := attr.FindVetoes(what).Check("DROP"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// Check the drop is not vetoed by the receiving inventory
	if veto := attr.FindVetoes(s.where.Parent()).Check("DROP"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// Get item's proper name
	name = attr.FindName(what).Name(name)

	// Move the item from our inventory to our location
	if from.Move(what, s.where) == nil {
		s.msg.Actor.SendBad("You cannot drop ", name, ".")
		return
	}

	who := attr.FindName(s.actor).Name("Someone")

	s.msg.Actor.SendGood("You drop ", name, ".")
	s.msg.Observer.SendInfo(who, " drops ", name, ".")
	s.ok = true
}
