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
	"strings"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
	"wolfmud.org/utils/inventory"
)

// direction type that can be easily change if needed
type direction uint8

// Constants for directions used for indexing. These constants can be used to
// index the directionNames array or the exits array in the Location struct
// using either the long or short constant name. For example:
//
// directionNames[location.S] is "South"
// l.directionalExits[location.South] retrieves the south exit for l
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

// directionNames are a map between a direction type and the textual name. This
// array can be indexed using the direction constants. For example:
//
//	directionNames[location.S] is "South"
//
var directionNames = [...]string{
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

// directionalExits hold the available directional exits from a location. There
// may be other exits implemented by things such as chutes or portals but these
// exits are not directional. The primary purpose of defining exits as a type is
// so that we can add a handy String method.
type directionalExits [len(directionNames)]Interface

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
func (e directionalExits) String() string {
	validExits := make([]string, 0, len(directionNames))
	for d, l := range e {
		if l != nil {
			validExits = append(validExits, directionNames[d])
		}
	}
	return strings.Join(validExits, ", ")
}

// Interface defines the methods for a basic location that all derived location
// types should implement.
type Interface interface {
	thing.Interface
	command.Interface
	inventory.Interface
	LinkExit(d direction, to Interface)
	Look(cmd *command.Command) (handled bool)
	Broadcast(omit []thing.Interface, format string, any ...interface{})
}

// Locateable defines the interface for something that has a location or can be
// moved from/to a location. For example a mobile.
type Locateable interface {
	Relocate(Interface) // Relocates a Locateable to a new location
	Locate() Interface  // Locate gets a Locateable's current location
}
