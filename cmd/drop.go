// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/text"
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

	// Check drop is not vetoed by item, item's current inventory or receiving
	// inventory
	for _, t := range []has.Thing{what, s.actor, s.where.Parent()} {
		for _, vetoes := range attr.FindAllVetoes(t) {
			if veto := vetoes.Check(s.actor, s.cmd); veto != nil {
				s.msg.Actor.SendBad(veto.Message())
				return
			}
		}
	}

	// Get item's proper name
	name = attr.FindName(what).Name(name)

	// Move the item from our inventory to our location
	from.Move(what, s.where)

	// As the Thing is now just laying on the ground check if it should register
	// for clean up
	attr.FindCleanup(what).Cleanup()

	// Re-enable actions if available
	attr.FindAction(what).Action()

	who := attr.FindName(s.actor).TheName("Someone")
	who = text.TitleFirst(who)

	s.msg.Actor.SendGood("You drop ", name, ".")
	s.msg.Observer.SendInfo(who, " drops ", name, ".")
	s.ok = true
}
