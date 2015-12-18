// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: QUIT
func init() {
	AddHandler(Quit, "QUIT")
}

// The Quit command acts as a hook for processing - such as cleanup - to be
// done when a player quits the game.
func Quit(s *state) {

	// Remove the player from the world
	if a := attr.FindLocate(s.actor); a != nil {
		if where := a.Where(); where != nil {
			where.Remove(s.actor)
		}
	}

	s.msg.actor.WriteString("\nBye bye...\n\n")
	s.ok = true
}
