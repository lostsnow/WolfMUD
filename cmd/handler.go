// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"
)

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
