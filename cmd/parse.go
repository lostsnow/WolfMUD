// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/has"
)

// Add handler for an empty command. The handler just acknowledges the empty
// command was processed by setting state.ok to true. We should not get empty
// commands from players as Parse screens them out. However other commands and
// possibly scripted commands might manually create a state accidentally with
// no command. Without this handler we would return the same as for an unknown
// or invalid command.
func init() {
	AddHandler(func(s *state) { s.ok = true }, "")
}

// Parse initiates processing of the input string for the specified Thing. The
// input string is expected to be either input from a player or possibly a scripted
// command. The actual command processed will be returned. For example GET or DROP.
func Parse(t has.Thing, input string) string {
	s := NewState(t, input)
	s.parse()
	return s.cmd
}
