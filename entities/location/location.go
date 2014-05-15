// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package location defines all of the different location types. Objects or
// mobiles can be 'at' a location if they are in the locations inventory.
// Locations are also the main locking mechanism in WolfMUD. By locking a
// location - usually indirectly via a Command - nothing else can touch the
// location or it's inventory while any changes are made. See the Command type
// for more details.
package location

import (
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/utils/command"
	"code.wolfmud.org/WolfMUD.git/utils/inventory"
	"code.wolfmud.org/WolfMUD.git/utils/messaging"
	"strings"
)

// direction type that can be easily change if needed
type direction uint8

// Constants for directions used for indexing. These constants can be used to
// index the directionLongNames or directionShortNames arrays or the exits array
// in the Location struct using either the long or short constant name. For
// example:
//
//	directionLongNames[location.S] is "South"
//	directionShortNames[location.S] is "S"
//	l.directionalExits[location.South] retrieves the south exit for l
//
const (
	N, NORTH direction = iota, iota
	NE, NORTHEAST
	E, EAST
	SE, SOUTHEAST
	S, SOUTH
	SW, SOUTHWEST
	W, WEST
	NW, NORTHWEST
	U, UP
	D, DOWN
)

// directionLongNames are a map between a direction type and the textual long
// name. This array can be indexed using the direction constants. For example:
//
//	directionLongNames[location.S] is "South"
//
var directionLongNames = [...]string{
	N:  "North",
	NE: "Northeast",
	E:  "East",
	SE: "Southeast",
	S:  "South",
	SW: "Southwest",
	W:  "West",
	NW: "Northwest",
	U:  "Up",
	D:  "Down",
}

// directionShortNames are a map between a direction type and the textual short
// name. This array can be indexed using the direction constants. For example:
//
//	directionShortNames[location.S] is "S"
//
var directionShortNames = [...]string{
	N:  "N",
	NE: "NE",
	E:  "E",
	SE: "SE",
	S:  "S",
	SW: "SW",
	W:  "W",
	NW: "NW",
	U:  "U",
	D:  "D",
}

// directionShortIndex are a map between a direction textual short name and a
// direction type. This map can be indexed using the direction short name. For
// example:
//
//	directionShortNames["S"] is location.S
//
var directionShortIndex = map[string]direction{
	"N":  N,
	"NE": NE,
	"E":  E,
	"SE": SE,
	"S":  S,
	"SW": SW,
	"W":  W,
	"NW": W,
	"U":  U,
	"D":  D,
}

// directionalExits hold the available directional exits from a location. There
// may be other exits implemented by things such as chutes or portals but these
// exits are not directional. The primary purpose of defining exits as a type is
// so that we can add a handy String method.
type directionalExits [len(directionLongNames)]Interface

// String returns the available directional exits from a location as a plain
// string with each direction separated by commas. For example:
//
//	"East, Southeast, South"
//
// TODO: Implement long exits
//
// TODO: Implement 'blocked' exits which are not described. For example if
// there is a door to the west and the exit 'west' should not be described
// unless the door is opened.
func (e directionalExits) String() (text string) {
	validExits := make([]string, 0, len(directionLongNames))
	for d, l := range e {
		if l != nil {
			validExits = append(validExits, directionLongNames[d])
		}
	}

	if len(validExits) == 0 {
		text = "[CYAN]You can see no immediate exits from here."
	} else {
		text = "[CYAN]You can see exits: [YELLOW]" + strings.Join(validExits, ", ")
	}

	return text
}

// Interface defines the methods for a basic location that all derived location
// types should implement.
type Interface interface {
	thing.Interface
	command.Interface
	inventory.Interface
	messaging.Broadcaster
	LinkExit(d direction, to Interface)
	look(cmd *command.Command) (handled bool)
	Lock()
	Unlock()
	Crowded() bool
}

// Locateable defines the interface for something that has a location or can be
// moved from/to a location. For example a mobile.
type Locateable interface {
	Relocate(Interface) // Relocates a Locateable to a new location
	Locate() Interface  // Locate gets a Locateable's current location
}
