// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: ( LOOK | L )
func init() {
	AddHandler(Look, "L", "LOOK")
}

func Look(s *state) {

	// Are we somewhere?
	if s.where == nil {
		s.msg.actor.WriteString("[ A Void ]\nYou are in a dark void. Around you nothing. No stars, no light, no heat and no sound.\n\nYou see no immediate exits from here.")
		return
	}

	what := s.where.Parent()

	s.msg.actor.WriteJoin("[ ", attr.FindName(what).Name("Somewhere"), " ]\n")

	mark := s.msg.actor.Len()

	for _, a := range attr.FindAllDescription(what) {
		s.msg.actor.WriteJoin(a.Description(), " ")
	}

	// If we added descriptions chop off space appended to last description
	// This is safe as ASCII space is only one byte
	if mark != s.msg.actor.Len() {
		s.msg.actor.Truncate(s.msg.actor.Len() - 1)
	}

	// Move off the current line and then write out a blank separator line
	s.msg.actor.WriteString("\n\n")
	mark = s.msg.actor.Len()

	if s.where.Crowded() {
		s.msg.actor.WriteJoin("You see a crowd here.\n")

		// NOTE: If location is crowded we don't list the items

	} else {

		// List mobiles here
		items := []has.Thing{}
		for _, c := range s.where.Contents() {

			if c == s.actor { // Don't include the looker in the list
				continue
			}

			if !attr.FindPlayer(c).Found() {
				items = append(items, c)
				continue
			}

			s.msg.actor.WriteJoin("You see ", attr.FindName(c).Name("someone"), " here.\n")
		}

		// List items here
		for _, i := range items {
			s.msg.actor.WriteJoin("You see ", attr.FindName(i).Name("something"), " here.\n")
		}
	}

	// If we wrote out any mobiles or items write out a blank separator line
	if mark != s.msg.actor.Len() {
		s.msg.actor.WriteString("\n")
	}

	s.msg.actor.WriteString(attr.FindExits(what).List())

	who := attr.FindName(s.actor).Name("Someone")
	s.msg.observer.WriteJoin(who, " starts looking around.")

	s.ok = true
}
