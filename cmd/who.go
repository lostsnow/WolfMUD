// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strconv"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/stats"
)

// Syntax: WHO
func init() {
	addHandler(who{}, "WHO")
}

type who cmd

func (who) process(s *state) {
	players := stats.List(s.actor)

	if len(players) == 0 {
		s.msg.Actor.SendInfo("You are all alone in this world.")
		return
	}

	for _, player := range players {
		s.msg.Actor.Send(player)
	}

	var (
		plural = len(players) > 1
		start  = "There is currently "
		end    = "."
	)

	if plural {
		start = "There are currently "
		end = "s."
	}

	s.msg.Actor.Send("")
	s.msg.Actor.Send(start, strconv.Itoa(len(players)), " other player", end)

	who := attr.FindName(s.actor).Name("Someone")
	s.msg.Observer.SendInfo("You see ", who, " concentrate for a moment.")

	s.ok = true
}
