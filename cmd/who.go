// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/stats"

	"strconv"
)

// Syntax: WHO
func init() {
	AddHandler(Who, "WHO")
}

func Who(s *state) {
	players := stats.List(s.actor)

	if len(players) == 0 {
		s.msg.Actor.WriteStrings("You are all alone in this world.")
		return
	}

	for _, player := range players {
		s.msg.Actor.WriteStrings(player, "\n")
	}

	var (
		plural = len(players) > 1
		start  = "\nThere is currently "
		end    = "."
	)

	if plural {
		start = "\nThere are currently "
		end = "s."
	}

	s.msg.Actor.WriteStrings(start, strconv.Itoa(len(players)), " other player", end)

	who := attr.FindName(s.actor).Name("Someone")
	s.msg.Observer.WriteStrings("You see ", who, " concentrate for a moment.")

	s.ok = true
}
