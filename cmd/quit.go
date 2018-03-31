// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/stats"
)

// Syntax: QUIT
func init() {
	addHandler(quit{}, "QUIT")
}

type quit cmd

// The Quit command acts as a hook for processing to be done when a player
// quits the game.
func (q quit) process(s *state) {

	// Instance of a junk command to handle locking and disposal
	j := junk{}

	// Make sure we are locking the reset origins for all of the player's
	// Inventory items so that they can be disposed of via junking.
	lc := len(s.locks)
	j.lockOrigins(s, s.actor)
	if len(s.locks) != lc {
		return
	}

	// Save player
	s.scriptActor("SAVE")

	// Junk everything we are carrying so it will either be reset or disposed of.
	// By calling junk.dispose directly we are bypassing all of the normal JUNK
	// command's Veto and Narrative checking. We also run silently without
	// generating any of the JUNK command's normal messages. This is deliberate
	// so that everything gets reset or disposed of when a player quits.
	for _, t := range attr.FindInventory(s.actor).Contents() {
		j.dispose(t)
	}

	// Reset the player's prompt while the Player is still in the Inventory we
	// are locking.
	attr.FindPlayer(s.actor).SetPromptStyle(has.StyleNone)

	// Remove the player from the world
	if s.where != nil {
		who := attr.FindName(s.actor).Name("someone")
		s.msg.Observer.SendInfo(who, " gives a strangled cry of 'Bye Bye', slowly fades away and is gone.")
		s.where.Disable(s.actor)
		s.where.Remove(s.actor)
	}

	s.msg.Actor.SendGood("You leave this world behind.")
	stats.Remove(s.actor)

	s.ok = true
}
