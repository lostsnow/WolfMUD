// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: HIT <who>
func init() {
	addHandler(hit{}, "HIT")
}

type hit cmd

func (hit) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to hit... someone?")
		return
	}

	if s.where.Crowded() {
		s.msg.Actor.SendInfo("It's too crowded in here to start a fight.")
		return
	}

	matches, words := Match(s.words, s.where.Everything())
	match := matches[0]
	mark := s.msg.Actor.Len()

	switch {
	case len(words) != 0: // Not exact match?
		name := strings.Join(s.words, " ")
		s.msg.Actor.SendBad("You see no '", name, "' to hit.")

	case len(matches) != 1: // More than one match?
		s.msg.Actor.SendBad("You can only hit one person at a time.")

	case match.Unknown != "":
		s.msg.Actor.SendBad("You see no '", match.Unknown, "' to hit.")

	case match.NotEnough != "":
		s.msg.Actor.SendBad("There are not that many '", match.NotEnough, "' to hit.")

	}

	// If we sent an error to the actor return now
	if mark != s.msg.Actor.Len() {
		return
	}

	s.participant = match.Thing

	who := attr.FindName(s.actor).TheName("Someone")
	what := attr.FindName(s.participant).TheName("Someone")

	h := attr.FindHealth(s.participant)

	if s.actor == s.participant {
		s.msg.Actor.SendGood("You give yourself a slap. Awake now?")
		s.msg.Observer.SendInfo(who, " slaps themself.")
		return
	}

	if !h.Found() {
		s.msg.Actor.SendBad("Hitting ", what, " is not going to accomplish much...")
		return
	}

	cur, _ := h.State()

	if cur <= 5 {
		s.msg.Actor.SendInfo(what, " already seems pretty beaten up.")
		s.ok = true
		return
	}

	h.Adjust(-5)

	s.msg.Actor.SendGood("You hit ", what)
	s.msg.Participant.SendBad(who, " hits you.")
	s.msg.Observer.SendInfo("You see ", who, " hit ", what, ".")

	s.ok = true
	return
}
