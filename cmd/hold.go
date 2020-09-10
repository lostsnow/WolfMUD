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

// Syntax: HOLD item...
func init() {
	addHandler(hold{}, "HOLD")
}

type hold cmd

func (hold) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to hold... something?")
		return
	}

	// Check actor has a non-empty inventory
	from := attr.FindInventory(s.actor)
	if from.Empty() {
		s.msg.Actor.SendBad("You don't have anything to hold.")
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
			s.msg.Actor.SendBad("You have no '", match.Unknown, "' to hold.")

		case match.NotEnough != "":
			s.msg.Actor.SendBad("You don't have that many '", match.NotEnough, "' to hold.")

		default:

			theName := attr.FindName(what).TheName("Something")

			// Is item holdable?
			h := attr.FindHoldable(what)
			if !h.Found() {
				s.msg.Actor.SendBad(text.TitleFirst(theName), " cannot be held.")
				return
			}

			body := attr.FindBody(s.actor)
			slots := h.Slots()

			// Is actor physically able to hold item? E.G. two hands?
			if !body.Found() || !body.Has(slots) {
				s.msg.Actor.SendBad("You can't physically hold ", theName, ".")
				continue nextMatch
			}

			// Is actor already using the item?
			if u := body.Usage(what); u != "" {
				s.msg.Actor.SendBad("You are already ", u, " ", theName, ".")
				continue nextMatch
			}

			// Check hold is not vetoed by item, item's current inventory
			for _, t := range []has.Thing{what, s.actor} {
				for _, vetoes := range attr.FindAllVetoes(t) {
					if veto := vetoes.Check(s.actor, s.cmd); veto != nil {
						s.msg.Actor.SendBad(veto.Message())
						continue nextMatch
					}
				}
			}

			if !body.Hold(h) {
				list := []string{}
				for _, t := range body.UsedBy(slots) {
					list = append(list, attr.FindName(t).Name("Something"))
				}
				s.msg.Actor.SendBad("You cannot hold ", theName, " while also using ", text.List(list), ".")
				continue nextMatch
			}

			s.msg.Actor.SendGood("You hold ", theName, ".")

			name := attr.FindName(what).Name("something")
			s.msg.Observer.SendInfo(who, " starts holding ", name, ".")
		}
	}

	s.ok = true
}
