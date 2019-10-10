// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: DROP item...
func init() {
	addHandler(drop{}, "DROP")
}

type drop cmd

func (drop) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to drop... something?")
		return
	}

	// Check actor has a non-empty inventory
	from := attr.FindInventory(s.actor)
	if from.Empty() {
		s.msg.Actor.SendBad("You don't have anything to drop.")
		return
	}

	who := attr.FindName(s.actor).TheName("Someone")
	who = text.TitleFirst(who)

	// Find matching items being carried
nextMatch:
	for _, match := range MatchAll(
		s.words,
		attr.FindInventory(s.actor).Contents(),
	) {
		what := match.Thing

		switch {
		case match.Unknown != "":
			s.msg.Actor.SendBad("You have no '", match.Unknown, "' to drop.")

		case match.NotEnough != "":
			s.msg.Actor.SendBad("You don't have that many '", match.NotEnough, "' to drop.")

		default:
			// Check drop is not vetoed by item, item's current inventory or receiving
			// inventory
			for _, t := range []has.Thing{what, s.actor, s.where.Parent()} {
				for _, vetoes := range attr.FindAllVetoes(t) {
					if veto := vetoes.Check(s.actor, s.cmd); veto != nil {
						s.msg.Actor.SendBad(veto.Message())
						continue nextMatch
					}
				}
			}

			theName := attr.FindName(what).TheName("something")

			// Move the item from actor's inventory to current location
			from.Move(what, s.where)

			// As the Thing is now just laying on the ground check if it should
			// register for clean up
			attr.FindCleanup(what).Cleanup()

			// Re-enable actions if available
			attr.FindAction(what).Action()

			s.msg.Actor.SendGood("You drop ", theName, ".")

			name := attr.FindName(what).Name("something")
			s.msg.Observer.SendInfo(who, " drops ", name, ".")
		}
	}

	s.ok = true
}
