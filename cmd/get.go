// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: GET item
func init() {
	addHandler(get{}, "GET")
}

type get cmd

func (get) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to get... something?")
		return
	}

	name := s.words[0]

	// Search for item we want to get in the inventory where we are
	what := s.where.Search(name)

	// Was item to get found?
	if what == nil {
		s.msg.Actor.SendBad("You see no '", name, "' to get.")
		return
	}

	// Check item not trying to get itself
	if what == s.actor {
		s.msg.Actor.SendInfo("Trying to pick youreself up by your bootlaces?")
		return
	}

	// Check we have an inventory so we can carry things
	to := attr.FindInventory(s.actor)
	if to == nil {
		s.msg.Actor.SendBad("You can't carry anything!")
		return
	}

	// Get item's proper name
	name = attr.FindName(what).Name(name)

	// Check the get is not vetoed by the item
	if veto := attr.FindVetoes(what).Check("GET"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// Check the get is not vetoed by the parent of the item's inventory
	if veto := attr.FindVetoes(s.where.Parent()).Check("GET"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// If item is a narrative we can't get it. We do this check after the veto
	// checks as the vetos could give us a better message/reson for not being
	// able to get the item.
	if attr.FindNarrative(what).Found() {
		s.msg.Actor.SendBad("For some reason you cannot get ", name, ".")
		return
	}

	// Cancel any Cleanup or Action events
	attr.FindCleanup(what).Abort()
	attr.FindAction(what).Abort()

	// Check if item respawns when picked up
	what = attr.FindReset(what).Spawn()

	// Move the item from our location to our inventory
	s.where.Move(what, to)

	who := attr.FindName(s.actor).Name("Someone")

	s.msg.Actor.SendGood("You get ", name, ".")
	s.msg.Observer.SendInfo("You see ", who, " get ", name, ".")
	s.ok = true
}
