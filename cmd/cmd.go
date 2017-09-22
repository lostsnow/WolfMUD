// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

// init adds a handler for the empty command. See the process method for
// details.
func init() {
	AddHandler(cmd{}, "")
}

// cmd is the default type used to build commands.
type cmd struct{}

// process implents a handler for the empty command. The handler just
// acknowledges the empty command was processed by setting state.ok to true. We
// should not get empty commands from players as Parse screens them out.
// However other commands, and possibly scripted commands, might manually
// create a state accidentally with no command. Without this handler we would
// return the same as for an unknown or invalid command.
func (cmd) process(s *state) {
	s.ok = true
}
