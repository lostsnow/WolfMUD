// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: OPEN <door>
func init() {
	addHandler(open{}, "OPEN")
}

type open cmd

func (open) process(s *state) {
	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("What did you want to open?")
		return
	}

	// Find matching door at location
	matches, words := Match(s.words, s.where.Everything())
	match := matches[0]
	mark := s.msg.Actor.Len()

	switch {
	case len(words) != 0: // Not exact match?
		name := strings.Join(s.words, " ")
		s.msg.Actor.SendBad("You see no '", name, "' here to open.")

	case len(matches) != 1: // More than one match?
		s.msg.Actor.SendBad("You can only open one thing at a time.")

	case match.Unknown != "":
		s.msg.Actor.SendBad("You see no '", match.Unknown, "' here to open.")

	case match.NotEnough != "":
		s.msg.Actor.SendBad("There are not that many '", match.NotEnough, "' here to open.")

	}

	// If we sent an error to the actor return now
	if mark != s.msg.Actor.Len() {
		return
	}

	from := s.where
	what := match.Thing
	name := attr.FindName(what).TheName("something") // Get item's proper name

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
		s.msg.Actor.SendGood("You open ", name, ".")

		who := attr.FindName(s.actor).TheName("Someone")
		name = attr.FindName(what).Name(name)
		s.msg.Observers[from].SendInfo(text.TitleFirst(who), " opens ", name, ".")
		s.msg.Observers[to].SendInfo(text.TitleFirst(name), " opens.")
	}

	s.ok = true
	return
}
