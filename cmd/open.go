// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: OPEN <door>
func init() {
	AddHandler(Open, "OPEN")
}

func Open(s *state) {
	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("What did you want to open?")
		return
	}

	name := s.words[0]
	from := s.where

	// Search for item to open in the inventory where we are
	what := s.where.Search(name)

	// Was item to get found?
	if what == nil {
		s.msg.Actor.SendBad("You see no '", name, "' to open.")
		return
	}

	// Get item's proper name
	name = attr.FindName(what).Name(name)

	// Is item a door that can be opened?
	door := attr.FindDoor(what)
	if !door.Found() {
		s.msg.Actor.SendBad("You cannot open ", name, ".")
		return
	}

	if door.Opened() {
		s.msg.Actor.SendInfo(text.TitleFirst(name), " is already open.")
		return
	}

	// Find out where the door leads to
	exits := attr.FindExits(from.Parent())
	to := exits.LeadsTo(door.Direction())

	// Are we locking where the door leads to yet? If not add it to the locks and
	// simply return. The parser will detect the locks have changed and reprocess
	// the command with the new locks held.
	if !s.CanLock(to) {
		s.AddLock(to)
		return
	}

	door.Open()

	if s.actor == what {
		s.msg.Observers[to].SendInfo(text.TitleFirst(name), " opens.")
		s.msg.Observers[from].SendInfo(text.TitleFirst(name), " opens.")
	} else {
		who := attr.FindName(s.actor).Name("Someone")
		s.msg.Actor.SendGood("You open ", name, ".")
		s.msg.Observers[from].SendInfo(who, " opens ", name, ".")
		s.msg.Observers[to].SendInfo(text.TitleFirst(name), " opens.")
	}

	s.ok = true
	return
}
