// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: ( N | NORTH | NE | NORTHEAST | E | EAST | SE | SOUTHEAST | S | SOUTH
//				 | SW | SOUTHWEST | W | WEST | NW | NORTHWEST | U | UP | D | DOWN)
//
// TODO: Move does not support vetoes yet.
func Move(s *state) {

	var from has.Inventory

	// Can we find out where we are?
	if a := attr.FindLocate(s.actor); a != nil {
		from = a.Where()
	}

	// A thing can only move itself if it knows where it is
	if from == nil {
		s.msg.actor.WriteString("You are not sure where you are, let alone where you are going!")
		return
	}

	// Is where we are exitable?
	exits := attr.FindExits(from.Parent())
	if exits == nil {
		s.msg.actor.WriteString("You can't see anywhere to go from here.")
		return
	}

	// Is direction a valid direction? Move could have been called directly by
	// another command just passing in the direction.
	direction := exits.NormalizeDirection(s.cmd)
	if direction == "" {
		s.msg.actor.WriteString("You wanted to go which way!?")
		return
	}

	// Find out where our exit leads to
	to := exits.LeadsTo(direction)
	if to == nil {
		s.msg.actor.WriteJoin("You can't go ", direction, " from here!")
		return
	}

	// Are we locking our destination yet? If not add it to the locks and simply
	// return. The parser will detect the locks have changed and reprocess the
	// command with the new locks held.
	if !s.CanLock(to) {
		s.AddLock(to)
		return
	}

	if from.Remove(s.actor) == nil {
		s.msg.actor.WriteString("Something stops you from leaving here!")
		return
	}

	to.Add(s.actor)

	// Describe our destination
	Look(s)
}
