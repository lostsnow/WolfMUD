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

// Syntax: ( LOOK | L )
func init() {
	AddHandler(Look, "L", "LOOK")
}

func Look(s *state) {

	// Are we somewhere?
	if s.where == nil {
		s.msg.Actor.Send(text.Cyan, "A Void", text.Reset, "\nYou are in a dark void. Around you nothing. No stars, no light, no heat and no sound.\n\n", text.Cyan, "You see no immediate exits from here.", text.Reset)
		return
	}

	what := s.where.Parent()

	// Write the location title
	s.msg.Actor.Send(text.Cyan, attr.FindName(what).Name("Somewhere"), text.Reset)
	s.msg.Actor.Send("")

	// Write the location descriptions
	for _, a := range attr.FindAllDescription(what) {
		s.msg.Actor.Append(a.Description())
	}

	// Write out a blank line and remember how many message we have sent so far
	s.msg.Actor.Send("")
	mark := s.msg.Actor.Len()

	// Write the location contents
	if s.where.Crowded() {
		s.msg.Actor.Send(text.Green, "You see a crowd here.")

		// NOTE: If location is crowded we don't list the items

	} else {

		// List mobiles here
		items := []has.Thing{}
		for _, c := range s.where.Contents() {

			// Don't include the actor doing the looking in the list
			if c == s.actor {
				continue
			}

			// If not a player it's an item so remember it instead of displaying it
			if !attr.FindPlayer(c).Found() {
				items = append(items, c)
				continue
			}

			s.msg.Actor.Send(text.Green, "You see ", attr.FindName(c).Name("someone"), " here.")
		}

		// Now write out the remembered items
		for _, i := range items {
			s.msg.Actor.Send(text.Yellow, "You see ", attr.FindName(i).Name("something"), " here.")
		}
	}

	// If we wrote any messages since the laste blank line out another blank
	// line. This prevents two blanks lines from being written if there is
	// nothing else here.
	if mark != s.msg.Actor.Len() {
		s.msg.Actor.Send("")
	}

	// Write out the exits
	s.msg.Actor.Send(text.Cyan, attr.FindExits(what).List())

	// Notify any observers we are looking around
	who := attr.FindName(s.actor).Name("Someone")
	s.msg.Observer.SendInfo(who, " starts looking around.")

	s.ok = true
}
