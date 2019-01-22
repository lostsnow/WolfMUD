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
	matches, unknowns, _ := MatchAll(
		s.words,
		attr.FindInventory(s.actor).Contents(),
		s.where.Everything(),
	)

	s.msg.Actor.SendGood("You look around.")
	for _, m := range matches {
		if attr.FindLocate(m).Where() == s.where {
			s.msg.Actor.Append("\nYou see ", attr.FindName(m).Name("something"), " here.")
		} else {
			s.msg.Actor.Append("\nYou are carrying ", attr.FindName(m).Name("something"), ".")
		}
	}

	if len(unknowns) > 0 {
		for _, unknown := range unknowns {
			s.msg.Actor.SendBad("You see no '", unknown, "' here.")
		}
	}

	s.ok = true
}
