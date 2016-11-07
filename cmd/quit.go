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
		if alias := attr.FindAlias(t); alias.Found() {
			aliases := alias.Aliases()
			s.scriptAll("DROP", aliases[0])
		}
	}

	// Remove the player from the world
	if s.where != nil {
		s.msg.observer.WriteStrings(who, " gives a strangled cry of 'Bye Bye', slowly fades away and is gone.")
		s.where.Remove(s.actor)
	}

	s.msg.actor.WriteString("You leave this world behind.")
	stats.Remove(s.actor)

	attr.FindPlayer(s.actor).SetPromptStyle(has.StyleNone)

	s.ok = true
}
