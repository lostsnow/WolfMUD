// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: ( LOOK | L )
func init() {
	AddHandler(Look, "L", "LOOK")
}

//
// BUG(diddymus): If we use LOOK and where we are has a narrative attribute but
// no inventory attribute the narrative contents will be listed by mistake. Is
// this an example of a bigger problem with narratives/inventories still?
func Look(s *state) {

	// Do we know where we are?
	where := s.where

	// Or are we the where?
	if where == nil {
		if a := attr.FindInventory(s.actor); a != nil {
			where = a
		}
	}

	// Still not anywhere?
	if where == nil {
		s.msg.actor.WriteString("You are in a dark void. Around you nothing. No stars, no light, no heat and no sound.")
		return
	}

	what := where.Parent()

	if a := attr.FindName(what); a != nil {
		s.msg.actor.WriteJoin("[ ", a.Name(), " ]\n")
	}

	mark := s.msg.actor.Len()

	for _, a := range attr.FindAllDescription(what) {
		s.msg.actor.WriteJoin(a.Description(), " ")
	}

	// If we added descriptions chop off space appended to last description
	// This is safe as ASCII space is only one byte
	if mark != s.msg.actor.Len() {
		s.msg.actor.Truncate(s.msg.actor.Len() - 1)
	}

	// Move off the current line and then write out a blank separator line
	s.msg.actor.WriteString("\n\n")
	mark = s.msg.actor.Len()

	if where.Crowded() {
		s.msg.actor.WriteJoin("You see a crowd here.\n")

		// NOTE: If location is crowded we don't list the items

	} else {

		// List mobiles here
		items := []has.Thing{}
		for _, c := range where.Contents() {

			if c == s.actor { // Don't include the looker in the list
				continue
			}

			if attr.FindPlayer(c) == nil {
				items = append(items, c)
				continue
			}

			if a := attr.FindName(c); a != nil {
				s.msg.actor.WriteJoin("You see ", a.Name(), " here.\n")
			}
		}

		// List items here
		for _, i := range items {
			if a := attr.FindName(i); a != nil {
				s.msg.actor.WriteJoin("You see ", a.Name(), " here.\n")
			}
		}
	}

	// If we wrote out any mobiles or items write out a blank separator line
	if mark != s.msg.actor.Len() {
		s.msg.actor.WriteString("\n")
	}

	if a := attr.FindExits(what); a != nil {
		s.msg.actor.WriteString(a.List())
	} else {
		s.msg.actor.WriteString("You see no immediate exits from here.")
	}

	s.ok = true
}
