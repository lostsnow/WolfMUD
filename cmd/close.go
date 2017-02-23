// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: CLOSE <door>
func init() {
	AddHandler(Close, "CLOSE")
}

func Close(s *state) {
	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("What did you want to close?")
		return
	}

	name := s.words[0]

	// Search for item to close in the inventory where we are
	what := s.where.Search(name)

	// Was item to get found?
	if what == nil {
		s.msg.Actor.SendBad("You see no '", name, "' to close.")
		return
	}

	// Get item's proper name
	name = attr.FindName(what).Name(name)

	// Is item a door that can be closed?
	door := attr.FindDoor(what)
	if !door.Found() {
		s.msg.Actor.SendBad("You cannot close ", name, ".")
		return
	}

	if door.Closed() {
		s.msg.Actor.SendInfo(text.TitleFirst(name), " is already closed.")
		return
	}

	door.Close()

	if s.actor == what {
		s.msg.Observer.SendInfo(text.TitleFirst(name), " closes.")
	} else {
		who := attr.FindName(s.actor).Name("Someone")
		s.msg.Actor.SendGood("You close ", name, ".")
		s.msg.Observer.SendInfo(who, " closes ", name, ".")
	}

	s.ok = true
	return
}
