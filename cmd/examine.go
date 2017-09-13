// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: ( EXAMINE | EXAM ) item
func init() {
	AddHandler(examine{}, "EXAM", "EXAMINE")
}

type examine cmd

func (examine) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You examine this and that, find nothing special.")
		return
	}

	name := s.words[0]

	// If we can, search where we are
	what := s.where.Search(name)

	// If item still not found try our own inventory
	if what == nil {
		what = attr.FindInventory(s.actor).Search(name)
	}

	// Was item to examine eventually found?
	if what == nil {
		s.msg.Actor.SendBad("You see no '", name, "' to examine.")
		return
	}

	// Check examine is not vetoed by item
	if veto := attr.FindVetoes(what).Check("EXAMINE", "EXAM"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// Get item's proper name
	name = attr.FindName(what).Name(name)

	s.msg.Actor.Send("You examine ", name, ".")

	for _, a := range attr.FindAllDescription(what) {
		s.msg.Actor.Append(a.Description())
	}

	// BUG(diddymus): If you examine another player you can see their inventory
	// items. For now we just skip the inventory listing if we are examining a
	// player.
	if !attr.FindPlayer(what).Found() {
		if l := attr.FindInventory(what).List(); l != "" {
			s.msg.Actor.Append(l)
		}
	}

	who := attr.FindName(s.actor).Name("Someone")

	s.msg.Observer.SendInfo(who, " studies ", name, ".")

	s.ok = true
}
