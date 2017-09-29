// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
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

	name := s.words[0]

	// Try searching inventory where we are
	what := s.where.Search(name)

	// If item still not found try our own inventory
	if what == nil {
		what = attr.FindInventory(s.actor).Search(name)
	}

	// Was item to read found?
	if what == nil {
		s.msg.Actor.SendBad("You see no '", name, "' to read.")
		return
	}

	// Get item's proper name
	name = attr.FindName(what).Name("something")

	// Find if item has writing
	writing := attr.FindWriting(what).Writing()

	// Was writing found?
	if writing == "" {
		s.msg.Actor.SendBad("You see no writing on ", name, " to read.")
		return
	}

	s.msg.Actor.Send("You read ", name, ". ", writing)

	who := attr.FindName(s.actor).Name("Someone")
	s.msg.Observer.SendInfo("You see ", who, " read ", name, ".")

	s.ok = true
}
