// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: ( N | NORTH | NE | NORTHEAST | E | EAST | SE | SOUTHEAST | S | SOUTH
//				 | SW | SOUTHWEST | W | WEST | NW | NORTHWEST | U | UP | D | DOWN)
//
func init() {
	addHandler(move{},
		"N", "NE", "E", "SE", "S", "SW", "W", "NW", "U", "D",
		"NORTH", "NORTHEAST", "EAST", "SOUTHEAST",
		"SOUTH", "SOUTHWEST", "WEST", "NORTHWEST",
		"UP", "DOWN",
	)
}

type move cmd

func (move) process(s *state) {

	from := s.where

	// Is where we are exitable?
	exits := attr.FindExits(from.Parent())
	if !exits.Found() {
		s.msg.Actor.SendInfo("You can't see anywhere to go from here.")
		return
	}

	// Is direction a valid direction? Move could have been called directly by
	// another command just passing in the direction.
	direction, err := exits.NormalizeDirection(s.cmd)
	if err != nil {
		s.msg.Actor.SendBad("You wanted to go which way!?")
		return
	}

	wayToGo := exits.ToName(direction)

	// Find out where our exit leads to
	to := exits.LeadsTo(direction)
	if to == nil {
		s.msg.Actor.SendBad("You can't go ", wayToGo, " from here!")
		return
	}

	// Are we locking our destination yet? If not add it to the locks and simply
	// return. The parser will detect the locks have changed and reprocess the
	// command with the new locks held.
	if !s.CanLock(to) {
		s.AddLock(to)
		return
	}

	// Check direction is not vetoed by any Thing here. The check can be
	// expensive so we do it after relocking the destination so it is only
	// performed once.
	canVeto := append(from.Narratives(), from.Contents()...)
	canVeto = append(canVeto, from.Parent())
	for _, t := range canVeto {
		for _, vetoes := range attr.FindAllVetoes(t) {
			if veto := vetoes.Check(s.actor, s.cmd); veto != nil {
				s.msg.Actor.SendBad(veto.Message())
				return
			}
		}
	}

	// Move us from where we are to our new location
	from.Move(s.actor, to)

	// Re-point where we are and re-alias observer
	s.where = to
	s.msg.Observer = s.msg.Observers[s.where]

	// Get actors name
	name := attr.FindName(s.actor).Name("someone")

	// Broadcast leaving and arrival notifications
	s.msg.Observers[from].SendInfo("You see ", name, " go ", wayToGo, ".")
	s.msg.Observers[to].SendInfo("You see ", name, " enter.")

	// Describe our destination
	s.scriptActor("LOOK")
}
