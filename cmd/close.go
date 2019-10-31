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

// Syntax: CLOSE <door>
func init() {
	addHandler(close{}, "CLOSE")
}

type close cmd

func (close) process(s *state) {
	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("What did you want to close?")
		return
	}

	name := strings.Join(s.words, " ")

	// Find matching door at location
	matches, words := Match(s.words, s.where.Everything())
	match := matches[0]
	mark := s.msg.Actor.Len()

	switch {
	case match.Unknown != "":
		s.msg.Actor.SendBad("You see no '", match.Unknown, "' here to close.")

	case match.NotEnough != "":
		s.msg.Actor.SendBad("There are not that many '", match.NotEnough, "' here to close.")

	case len(words) != 0: // Not exact match?
		s.msg.Actor.SendBad("You see no '", name, "' here to close.")

	case len(matches) != 1: // More than one match?
		s.msg.Actor.SendBad("You can only close one thing at a time.")

	}

	// If we sent an error to the actor return now
	if mark != s.msg.Actor.Len() {
		return
	}

	from := s.where
	what := match.Thing
	name = attr.FindName(what).TheName(name) // Get item's proper name

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
		s.msg.Actor.SendGood("You close ", name, ".")

		who := attr.FindName(s.actor).TheName("Someone")
		name = attr.FindName(what).Name(name)
		s.msg.Observers[from].SendInfo(text.TitleFirst(who), " closes ", name, ".")
		s.msg.Observers[to].SendInfo(text.TitleFirst(name), " closes.")
	}

	s.ok = true
	return
}
