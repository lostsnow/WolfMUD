// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/has"

	"strings"
)

func Parse(t has.Thing, input string) (msg string, ok bool) {
	input = strings.ToUpper(input)
	words := strings.Fields(input)

	if len(words) == 0 {
		return
	}

	cmd := words[0]

	switch cmd {
	case "N", "NE", "E", "SE", "S", "SW", "W", "NW", "U", "D":
		return Move(t, cmd)
	case "NORTH", "NORTHEAST", "EAST", "SOUTHEAST", "SOUTH", "SOUTHWEST", "WEST", "NORTHWEST", "UP", "DOWN":
		return Move(t, cmd)
	case "#DUMP":
		return Dump(t, words[1:])
	case "DROP":
		return Drop(t, words[1:])
	case "EXAMINE", "EXAM":
		return Examine(t, words[1:])
	case "GET":
		return Get(t, words[1:])
	case "INVENTORY", "INV":
		return Inventory(t)
	case "LOOK", "L":
		return Look(t)
	case "PUT":
		return Put(t, words[1:])
	case "READ":
		return Read(t, words[1:])
	case "TAKE":
		return Take(t, words[1:])
	case "VERSION":
		return Version(t)
	default:
		return "Eh?", false
	}
}
