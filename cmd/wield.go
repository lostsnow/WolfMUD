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

// Syntax: WIELD item...
func init() {
	addHandler(wield{}, "WIELD")
}

type wield cmd

func (wield) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to wield... something?")
		return
	}

	// Check actor has a non-empty inventory
	from := attr.FindInventory(s.actor)
	if from.Empty() {
		s.msg.Actor.SendBad("You don't have anything to wield.")
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
			s.msg.Actor.SendBad("You have no '", match.Unknown, "' to wield.")

		case match.NotEnough != "":
			s.msg.Actor.SendBad("You don't have that many '", match.NotEnough, "' to wield.")

		default:

			theName := attr.FindName(what).TheName("Something")

			// Is item wieldable?
			w := attr.FindWieldable(what)
			if !w.Found() {
				s.msg.Actor.SendBad(text.TitleFirst(theName), " cannot be wielded.")
				return
			}

			body := attr.FindBody(s.actor)
			slots := w.Slots()

			// Is actor physically able to wield item? E.G. two hands?
			if !body.Found() || !body.Has(slots) {
				s.msg.Actor.SendBad("You can't physically wield ", theName, ".")
				continue nextMatch
			}

			// Is actor already using the item?
			if u := body.Usage(what); u != "" {
				s.msg.Actor.SendBad("You are already ", u, " ", theName, ".")
				continue nextMatch
			}

			// Check wield is not vetoed by item, item's current inventory
			for _, t := range []has.Thing{what, s.actor} {
				for _, vetoes := range attr.FindAllVetoes(t) {
					if veto := vetoes.Check(s.actor, s.cmd); veto != nil {
						s.msg.Actor.SendBad(veto.Message())
						continue nextMatch
					}
				}
			}

			if !body.Wield(w) {
				list := []string{}
				for _, t := range body.UsedBy(slots) {
					list = append(list, attr.FindName(t).Name("Something"))
				}
				s.msg.Actor.SendBad("You cannot wield ", theName, " while also using ", text.List(list), ".")
				continue nextMatch
			}

			s.msg.Actor.SendGood("You wield ", theName, ".")

			name := attr.FindName(what).Name("something")
			s.msg.Observer.SendInfo(who, " starts wielding ", name, ".")
		}
	}

	s.ok = true
}
