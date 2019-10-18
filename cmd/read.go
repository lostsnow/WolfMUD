// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: READ item
func init() {
	addHandler(read{}, "READ")
}

type read cmd

func (read) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("Did you want to read something specific?")
		return
	}

	// Find matching item at location or held by actor
	matches, words := Match(
		s.words,
		s.where.Everything(),
		attr.FindInventory(s.actor).Contents(),
	)
	match := matches[0]
	mark := s.msg.Actor.Len()

	switch {
	case len(words) != 0: // Not exact match?
		name := strings.Join(s.words, " ")
		s.msg.Actor.SendBad("You see no '", name, "' to read.")

	case len(matches) != 1: // More than one match?
		s.msg.Actor.SendBad("You can only read one thing at a time.")

	case match.Unknown != "":
		s.msg.Actor.SendBad("You see no '", match.Unknown, "' read.")

	case match.NotEnough != "":
		s.msg.Actor.SendBad("There are not that many '", match.NotEnough, "' to read.")

	}

	// If we sent an error to the actor return now
	if mark != s.msg.Actor.Len() {
		return
	}

	what := match.Thing
	name := attr.FindName(what).TheName("something") // Get item's proper name

	// Find if item has writing
	writing := attr.FindWriting(what).Writing()

	// Was writing found?
	if writing == "" {
		s.msg.Actor.SendBad("You see no writing on ", name, " to read.")
		return
	}

	s.msg.Actor.Send("You read ", name, ". ", writing)

	who := attr.FindName(s.actor).Name("Someone")
	name = attr.FindName(what).Name("something")
	s.msg.Observer.SendInfo("You see ", who, " read ", name, ".")

	s.ok = true
}
