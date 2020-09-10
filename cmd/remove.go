// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: REMOVE item...
func init() {
	addHandler(remove{}, "REMOVE")
}

type remove cmd

func (remove) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to remove... something?")
		return
	}

	// Check actor has a non-empty inventory
	from := attr.FindInventory(s.actor)
	if from.Empty() {
		s.msg.Actor.SendBad("You don't have anything to remove.")
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
			s.msg.Actor.SendBad("You have no '", match.Unknown, "' to remove.")

		case match.NotEnough != "":
			s.msg.Actor.SendBad("You don't have that many '", match.NotEnough, "' to remove.")

		default:

			theName := attr.FindName(what).TheName("Something")
			b := attr.FindBody(s.actor)

			// Is item being used?
			if !b.Using(what) {
				s.msg.Actor.SendBad("You are not currently using ", theName, ".")
				continue nextMatch
			}

			// Check remove is not vetoed by item, item's current inventory
			for _, t := range []has.Thing{what, s.actor} {
				for _, vetoes := range attr.FindAllVetoes(t) {
					if veto := vetoes.Check(s.actor, s.cmd); veto != nil {
						s.msg.Actor.SendBad(veto.Message())
						continue nextMatch
					}
				}
			}

			u := b.Usage(what)
			b.Remove(what)
			s.msg.Actor.SendGood("You stop ", u, " ", theName, ".")

			name := attr.FindName(what).Name("something")
			s.msg.Observer.SendInfo(who, " stops ", u, " ", name, ".")
		}
	}

	s.ok = true
}
