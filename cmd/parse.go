// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/has"

	"strings"
	"unicode"
)

// Parse initiates processing of the input string for the specified Thing. The
// input string is expected to be either a players input or possibly a scripted
// command. It returns msg which is the response to carry out the command and
// ok which is set to true if the command completed successfully else false.
// The ok flag can be used by other commands and possibly scripting to check if
// the command was successful without having to try and parse the msg.
func Parse(t has.Thing, input string) (msg string, ok bool) {

	if strings.TrimLeftFunc(input, unicode.IsSpace) == "" {
		return
	}

	s := NewState(t, input)
	s.parse(dispatch)

	return s.msg.actor.String(), s.ok
}

var handlers = map[string]func(*state){
	"N":  Move,
	"NE": Move,
	"E":  Move,
	"SE": Move,
	"S":  Move,
	"SW": Move,
	"W":  Move,
	"NW": Move,
	"U":  Move,
	"D":  Move,

	"NORTH":     Move,
	"NORTHEAST": Move,
	"EAST":      Move,
	"SOUTHEAST": Move,
	"SOUTH":     Move,
	"SOUTHWEST": Move,
	"WEST":      Move,
	"NORTHWEST": Move,
	"UP":        Move,
	"DOWN":      Move,

	"#DUMP":     Dump,
	"DROP":      Drop,
	"EXAM":      Examine,
	"EXAMINE":   Examine,
	"GET":       Get,
	"INV":       Inventory,
	"INVENTORY": Inventory,
	"L":         Look,
	"LOOK":      Look,
	"PUT":       Put,
	"QUIT":      Quit,
	"READ":      Read,
	"TAKE":      Take,
	"VERSION":   Version,
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
