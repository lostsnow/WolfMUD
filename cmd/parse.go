// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/has"

	"errors"
)

// EndOfDataError represents the fact that no more data is expected to be
// returned. For example the QUIT command has been used.
var (
	EndOfDataError = errors.New("End of data - player quitting")
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
// input string is expected to be either a players input or possibly a scripted
// command. If there is to be no more data - for example because the QUIT
// command has been issued - an EndOfDataError will be returned. Otherwise nil
// is returned.
func Parse(t has.Thing, input string) error {
	s := NewState(t, input)
	s.parse(dispatch)

	// If actor is quitting call Close to deallocate all of the Thing Attribute.
	// We can do that at this point as all message buffers have been written, all
	// locks released and the actor removed from the world.
	if string(s.cmd) == "QUIT" {
		s.actor.Close()
		return EndOfDataError
	}

	return nil
}
