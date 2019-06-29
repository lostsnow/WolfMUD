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

// Syntax: GET item...
func init() {
	addHandler(get{}, "GET")
}

type get cmd

func (get) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to get... something?")
		return
	}

	// Check actor has an inventory to put things into
	to := attr.FindInventory(s.actor)
	if !to.Found() {
		s.msg.Actor.SendBad("You can't carry anything!")
		return
	}

	who := attr.FindName(s.actor).TheName("Someone")

	// Find matching items at location
nextMatch:
	for _, match := range MatchAll(
		s.words,
		s.where.Everything(),
	) {
		what := match.Thing

		switch {
		case match.Unknown != "":
			s.msg.Actor.SendBad("You see no '", match.Unknown, "' to get.")

		case match.NotEnough != "":
			s.msg.Actor.SendBad("You don't see that many '", match.NotEnough, "' to get.")

		case what == s.actor:
			s.msg.Actor.SendInfo("Trying to pick youreself up by your bootlaces?")

		default:
			// Check get is not vetoed by item, item's current inventory or receiving
			// inventory
			for _, t := range []has.Thing{what, s.where.Parent(), s.actor} {
				for _, vetoes := range attr.FindAllVetoes(t) {
					if veto := vetoes.Check(s.actor, s.cmd); veto != nil {
						s.msg.Actor.SendBad(veto.Message())
						continue nextMatch
					}
				}
			}

			nameAttr := attr.FindName(what)

			// If item is a player we don't allow them to be picked up. This solves a
			// number of logistical problems - like the player quitting while held or
			// being put into a carried container.
			//
			// BUG(diddymus): It should be possible to put a GET Veto on a player.
			// However due to the way GET applies Vetoes we can't yet.
			if attr.FindPlayer(what).Found() {
				name := text.TitleFirst(nameAttr.TheName("someone"))
				s.msg.Actor.SendBad(name, " does not want to be picked up!")
				continue nextMatch
			}

			name := nameAttr.Name("something")

			// If item is a narrative we can't get it. We do this check after the
			// veto checks as the vetos could give us a better message/reson for not
			// being able to get the item.
			if attr.FindNarrative(what).Found() {
				s.msg.Actor.SendBad("For some reason you cannot get ", name, ".")
				continue nextMatch
			}

			// Cancel any pending Cleanup or Action events
			attr.FindCleanup(what).Abort()
			attr.FindAction(what).Abort()

			// If item respawns when picked up take newly spawned copy
			if s := attr.FindReset(what).Spawn(); s != nil {
				what = s
			}

			// Move the item from current location to actor's inventory
			s.where.Move(what, to)

			s.msg.Actor.SendGood("You get ", name, ".")
			s.msg.Observer.SendInfo("You see ", who, " get ", name, ".")
		}
	}

	s.ok = true
}
