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

func dispatch(s *state) {

	var (
		msg string
		ok  bool

		cmd   = s.cmd
		words = s.words
		t     = s.actor
	)

	// Dummy usage to avoid 'declared and not used'
	// compile errors as we incrementally clean up
	// the code in this function while converting
	// commands to use parser state
	_ = words
	_ = t
	_ = cmd
	_ = msg
	_ = ok

	switch cmd {
	case "N", "NE", "E", "SE", "S", "SW", "W", "NW", "U", "D":
		msg, ok = Move(t, cmd)
	case "NORTH", "NORTHEAST", "EAST", "SOUTHEAST", "SOUTH", "SOUTHWEST", "WEST", "NORTHWEST", "UP", "DOWN":
		msg, ok = Move(t, cmd)
	case "#DUMP":
		Dump(s)
	case "DROP":
		Drop(s)
	case "EXAMINE", "EXAM":
		Examine(s)
	case "GET":
		Get(s)
	case "INVENTORY", "INV":
		Inventory(s)
	case "LOOK", "L":
		Look(s)
	case "PUT":
		Put(s)
	case "QUIT":
		Quit(s)
	case "READ":
		msg, ok = Read(t, words[0:])
	case "TAKE":
		msg, ok = Take(t, words[0:])
	case "VERSION":
		Version(s)
	default:
		msg, ok = "Eh?", false
	}

	if msg != "" {
		s.msg.actor.WriteString(msg)
		s.ok = ok
	}
}
