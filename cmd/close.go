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
	from := s.where

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

	door.Close()

	if s.actor == what {
		s.msg.Observers[to].SendInfo(text.TitleFirst(name), " closes.")
		s.msg.Observers[from].SendInfo(text.TitleFirst(name), " closes.")
	} else {
		who := attr.FindName(s.actor).Name("Someone")
		s.msg.Actor.SendGood("You close ", name, ".")
		s.msg.Observers[from].SendInfo(who, " closes ", name, ".")
		s.msg.Observers[to].SendInfo(text.TitleFirst(name), " closes.")
	}

	s.ok = true
	return
}
