// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: WHICH item...
func init() {
	addHandler(which{}, "WHICH")
}

type which cmd

func (w which) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendGood("You look around for nothing in particular.")
		return
	}

	// Find items either being carried or at location
	for _, match := range MatchAll(
		s.words,
		attr.FindInventory(s.actor).Contents(),
		s.where.Everything(),
	) {
		switch {
		case match.Unknown != "":
			s.msg.Actor.SendBad("You see no '", match.Unknown, "' here.")
			continue

		case match.NotEnough != "":
			s.msg.Actor.SendBad("You don't see that many '", match.NotEnough, "' here.")
			continue

		default:
			if attr.FindLocate(match).Where() == s.where {
				s.msg.Actor.SendGood(
					"You see ", attr.FindName(match).Name("something"), " here.",
				)
			} else {
				s.msg.Actor.SendGood(
					"You are carrying ", attr.FindName(match).Name("something"), ".",
				)
			}
		}

	}

	s.ok = true
}
