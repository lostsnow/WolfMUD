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
		s.msg.Actor.WriteStrings("[ A Void ]\nYou are in a dark void. Around you nothing. No stars, no light, no heat and no sound.\n\nYou see no immediate exits from here.")
		return
	}

	what := s.where.Parent()

	s.msg.Actor.WriteStrings("[ ", attr.FindName(what).Name("Somewhere"), " ]")

	for x, a := range attr.FindAllDescription(what) {
		if x == 0 {
			s.msg.Actor.WriteStrings(a.Description())
		} else {
			s.msg.Actor.WriteAppend(a.Description())
		}
	}

	// Move off the current line and then write out a blank separator line
	s.msg.Actor.WriteStrings("")
	mark := s.msg.Actor.Len()

	if s.where.Crowded() {
		s.msg.Actor.WriteStrings("You see a crowd here.")

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

			s.msg.Actor.WriteStrings("You see ", attr.FindName(c).Name("someone"), " here.")
		}

		// List items here
		for _, i := range items {
			s.msg.Actor.WriteStrings("You see ", attr.FindName(i).Name("something"), " here.")
		}
	}

	// If we wrote out any mobiles or items write out a blank separator line
	if mark != s.msg.Actor.Len() {
		s.msg.Actor.WriteStrings("")
	}

	s.msg.Actor.WriteStrings(attr.FindExits(what).List())

	who := attr.FindName(s.actor).Name("Someone")
	s.msg.Observer.WriteStrings(who, " starts looking around.")

	s.ok = true
}
