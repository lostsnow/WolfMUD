// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

// Syntax: QUIT
func init() {
	AddHandler(Quit, "QUIT")
}

//
// The Quit command acts as a hook for processing - such as cleanup - to be
// done when a player quits the game.
func Quit(s *state) {
	s.msg.actor.WriteString("Bye bye...")
	s.ok = true
}
