// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"
)

// handler is the interface for command processing handlers.
type handler interface {
	process(*state)
}

// handlers is a list of commands and their handlers. AddHandler should be used
// to add new handlers. state.handleCommand uses this list to lookup the
// correct handler to invoke for a given command.
var handlers = map[string]handler{}

// AddHandler adds the given commands for the specified handler. The commands
// will automatically be uppercased. Each command and its aliases should
// register its handler in its init function. For example:
//
//	func init() {
//		AddHandler(Look{}, "L", "LOOK")
//	}
//
// In this example the LOOK command and it's alias 'L' register an instance of
// Look as their handler. If a handler is added for an existing command or
// alias the original handler will be replaced.
func AddHandler(h handler, cmd ...string) {
	for _, cmd := range cmd {
		handlers[strings.ToUpper(cmd)] = h
	}
}
