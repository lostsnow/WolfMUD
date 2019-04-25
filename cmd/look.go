// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: ( LOOK | L )
func init() {
	addHandler(look{}, "L", "LOOK")
}

type look cmd

func (look) process(s *state) {

	what := s.where.Parent()

	// Write the location title
	s.msg.Actor.Send(text.Cyan, text.TitleFirst(attr.FindName(what).Name("Somewhere")), text.Reset)
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
		// NOTE: If location is crowded we don't list the players or items
	} else {

		// List players here - but don't include the actor
		for _, p := range s.where.Players() {
			if p == s.actor {
				continue
			}
			s.msg.Actor.Send(text.Green, "You see ", attr.FindName(p).Name("someone"), " here.")
		}

		// List items here
		for _, c := range s.where.Contents() {
			s.msg.Actor.Send(text.Yellow, "You see ", attr.FindName(c).Name("something"), " here.")
		}
	}

	// If we wrote any messages since the last blank line out another blank line.
	// This prevents two blanks lines from being written if there is nothing else
	// here.
	if mark != s.msg.Actor.Len() {
		s.msg.Actor.Send("")
	}

	// Write out the exits
	s.msg.Actor.Send(text.Cyan, attr.FindExits(what).List())

	// Notify any observers we are looking around
	who := attr.FindName(s.actor).Name("Someone")
	who = text.TitleFirst(who)
	s.msg.Observer.SendInfo(who, " starts looking around.")

	s.ok = true
}
