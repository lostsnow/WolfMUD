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
	AddHandler(Examine, "EXAM", "EXAMINE")
}

func Examine(s *state) {

	if len(s.words) == 0 {
		s.msg.actor.WriteString("You examine this and that, find nothing special.")
		return
	}

	name := s.words[0]

	// If we can, search where we are
	what := s.where.Search(name)

	// If item still not found see if we can search narratives
	if what == nil && s.where != nil {
		if a := attr.FindNarrative(s.where.Parent()); a != nil {
			what = a.Search(name)
		}
	}

	// If item still not found try our own inventory
	if what == nil {
		what = attr.FindInventory(s.actor).Search(name)
	}

	// Was item to examine eventually found?
	if what == nil {
		s.msg.actor.WriteJoin("You see no '", name, "' to examine.")
		return
	}

	// Check examine is not vetoed by item
	if vetoes := attr.FindVetoes(what); vetoes != nil {
		if veto := vetoes.Check("EXAMINE"); veto != nil {
			s.msg.actor.WriteString(veto.Message())
			return
		}
	}

	// Get item's proper name
	name = attr.FindName(what).Name(name)

	s.msg.actor.WriteJoin("You examine ", name, ".")

	for _, d := range attr.FindAllDescription(what) {
		s.msg.actor.WriteJoin(" ", d.Description())
	}

	if l := attr.FindInventory(what).List(); l != "" {
		s.msg.actor.WriteJoin(" ", l)
	}

	s.ok = true
}
