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
//
// BUG(diddymus): If we use LOOK and where we are has a narrative attribute but
// no inventory attribute the narrative contents will be listed by mistake. Is
// this an example of a bigger problem with narratives/inventories still?
func Look(s *state) {

	// Do we know where we are?
	var where has.Inventory
	locater := attr.FindLocate(s.actor)
	if locater != nil {
		where = locater.Where()
	}

	// Or are we the where?
	if locater == nil {
		if a := attr.FindInventory(s.actor); a != nil {
			where = a
		}
	}

	// Still not anywhere?
	if where == nil {
		s.msg.actor.WriteString("You are in a dark void. Around you nothing. No stars, no light, no heat and no sound.")
		return
	}

	if a := attr.FindName(where.Parent()); a != nil {
		s.msg.actor.WriteJoin("[ ", a.Name(), " ]\n")
	}

	mark := s.msg.actor.Len()

	for _, d := range attr.FindAllDescription(where.Parent()) {
		s.msg.actor.WriteJoin(d.Description(), " ")
	}

	// If we added descriptions chop off space appended to last description
	// This is safe as ASCII space is only one byte
	if mark != s.msg.actor.Len() {
		s.msg.actor.Truncate(s.msg.actor.Len() - 1)
	}

	s.msg.actor.WriteString("\n\n")
	mark = s.msg.actor.Len()

	// Note: We don't want to include the looker in the list of things here which
	// is what the l != t check is for
	if a := attr.FindInventory(where.Parent()); a != nil {
		for _, l := range a.Contents() {
			if l == s.actor {
				continue
			}
			if n := attr.FindName(l); n != nil {
				s.msg.actor.WriteJoin("You can see ", n.Name(), " here.\n")
			}
		}
	}

	if mark != s.msg.actor.Len() {
		s.msg.actor.WriteString("\n")
	}

	if a := attr.FindExits(where.Parent()); a != nil {
		s.msg.actor.WriteString(a.List())
	} else {
		s.msg.actor.WriteString("You can see no immediate exits from here.")
	}

	s.ok = true
}
