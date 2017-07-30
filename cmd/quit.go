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
	AddHandler(Quit, "QUIT")
}

// The Quit command acts as a hook for processing - such as cleanup - to be
// done when a player quits the game.
func Quit(s *state) {

	who := attr.FindName(s.actor).Name("someone")

	// Drop any items we are carrying.
	//
	// NOTE: In future this needs to be updated to only drop temporary items.
	from := attr.FindInventory(s.actor)
	for _, t := range from.Contents() {
		s.scriptAll("DROP", t.UID())
	}

	// Reset the player's prompt while the Player is still in the Inventory we
	// are locking.
	attr.FindPlayer(s.actor).SetPromptStyle(has.StyleNone)

	// Remove the player from the world
	if s.where != nil {
		s.msg.Observer.SendInfo(who, " gives a strangled cry of 'Bye Bye', slowly fades away and is gone.")
		s.where.Remove(s.actor)
	}

	s.msg.Actor.SendGood("You leave this world behind.")
	stats.Remove(s.actor)

	s.ok = true
}
