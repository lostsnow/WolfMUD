// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: GET item
func init() {
	AddHandler(Get, "GET")
}

func Get(s *state) {

	if len(s.words) == 0 {
		s.msg.actor.WriteString("You go to get... something?")
		return
	}

	var (
		name = s.words[0]

		what        has.Thing
		isNarrative bool
	)

	// Are we somewhere?
	if s.where == nil {
		s.msg.actor.WriteString("You cannot get anything here.")
		return
	}

	// Search for item we want to get in the inventory where we are
	what = s.where.Search(name)

	// If item not found in inventory also check narratives where we are
	if what == nil {
		what = attr.FindNarrative(s.where.Parent()).Search(name)
		isNarrative = (what != nil)
	}

	// Was item to get found?
	if what == nil {
		s.msg.actor.WriteJoin("You see no '", name, "' to get.")
		return
	}

	// Check item not trying to get itself
	if what == s.actor {
		s.msg.actor.WriteString("Trying to pick youreself up by your bootlaces?")
		return
	}

	// Check we have an inventory so we can carry things
	to := attr.FindInventory(s.actor)
	if to == nil {
		s.msg.actor.WriteString("You can't carry anything!")
		return
	}

	// Get item's proper name
	name = attr.FindName(what).Name(name)

	// Check the get is not vetoed by the item
	if vetoes := attr.FindVetoes(what); vetoes != nil {
		if veto := vetoes.Check("GET"); veto != nil {
			s.msg.actor.WriteString(veto.Message())
			return
		}
	}

	// Check the get is not vetoed by the parent of the item's inventory
	if vetoes := attr.FindVetoes(s.where.Parent()); vetoes != nil {
		if veto := vetoes.Check("GET"); veto != nil {
			s.msg.actor.WriteString(veto.Message())
			return
		}
	}

	// If item is a narrative we can't get it. We do this check after the veto
	// checks as the vetos could give us a better message/reson for not being
	// able to take the item.
	if isNarrative {
		s.msg.actor.WriteJoin("You cannot get ", name, ".")
		return
	}

	// If all seems okay try and remove item from where it is
	if s.where.Remove(what) == nil {
		s.msg.actor.WriteJoin("You cannot get ", name, ".")
		return
	}

	// Add item to our inventory
	to.Add(what)

	s.msg.actor.WriteJoin("You get ", name, ".")
	s.ok = true
}
