// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/has"

	"strings"
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

// handlers is a list of commands and their handlers. AddHandler should be used
// to add new handlers. parser.dispatch uses this list to lookup the correct
// handler to invoke for a given command.
var handlers = map[string]func(*state){}

// AddHandler adds the given commands for the specified handler. The commands
// will automatically be uppercased. Each command and it's aliases should
// register it's handler in it's init function. For example:
//
//	func init() {
//		AddHandler(Look, "L", "LOOK")
//	}
//
// In this example the LOOK command and it's alias 'L' register the Look
// function as their handler. If a handler is added for an existing command or
// alias the original handler will be replaced.
func AddHandler(handler func(*state), cmd ...string) {
	for _, cmd := range cmd {
		handlers[strings.ToUpper(cmd)] = handler
	}
}

// Parse initiates processing of the input string for the specified Thing. The
// input string is expected to be either a players input or possibly a scripted
// command.
func Parse(t has.Thing, input string) {
	s := NewState(t, input)
	s.parse(dispatch)

	// If actor is quitting call Close to deallocate all of the Thing Attribute.
	// We can do that at this point as all message buffers have been written, all
	// locks released and the actor removed from the world.
	if string(s.cmd) == "QUIT" {
		s.actor.Close()
	}
}

// dispatch invokes the handler for a given command. The command is specified
// by state.cmd which is used to lookup the handler to invoke. dispatch should
// only be called once any required locks are held. This is usually taken care
// of by the state.parse method which is normally called by cmd.Parse.
func dispatch(s *state) {

	handler, valid := handlers[s.cmd]

	// Respond to an invalid command
	if !valid {
		s.msg.actor.WriteString("Eh?")
		return
	}
	handler(s)
}
