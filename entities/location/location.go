// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package location defines all of the location types. Objects or mobiles can
// be 'at' a location if they are in the locations inventory. Locations are
// also the main locking mechanism in WolfMUD. By locking a location - usually
// indirectly via a Command - nothing else can touch the location or it's
// inventory which any changes are made. See the Command type for more details.
package location

import (
	"fmt"
	"strings"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
	"wolfmud.org/utils/inventory"
	"wolfmud.org/utils/responder"
)

// direction type that can be easily change if needed
type direction uint8

// Constants for directions for indexing. These constants can be used to index
// the directionNames array or the exits array in the Location struct using
// either the long or short constant name. For example:
//
//	directionNames[location.S] is "South"
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

// Location provides a default location implementation
type Location struct {
	*thing.Thing
	*inventory.Inventory
	directionalExits
}

// New creates a new Location and returns a reference to it.
//
// NOTE: We could save memory at the cost of performance by not allocating the
// Inventory until something is added - via Add. We could also set it to nil
// when the last Thing is removed - via Remove. Performance wise we would incur
// a penality creating the Inventory and also create a lot more for the GC to
// handle?
func New(name string, aliases []string, description string) *Location {
	return &Location{
		Thing:     thing.New(name, aliases, description),
		Inventory: &inventory.Inventory{},
	}
}

// LinkExit links one location to another in the direction given. This is
// normally only done at setup time when the world is initially loaded.
//
// NOTE: The Java version had softlinking - is it still needed?
func (l *Location) LinkExit(d direction, to Interface) {
	l.directionalExits[d] = to
}

// Add puts a Thing at this location.
func (l *Location) Add(thing thing.Interface) {
	if t, ok := thing.(Locateable); ok {
		t.Relocate(l)
	}
	l.Inventory.Add(thing)
}

// Remove takes a Thing from this location.
func (l *Location) Remove(thing thing.Interface) {
	if t, ok := thing.(Locateable); ok {
		t.Relocate(nil)
	}
	l.Inventory.Remove(thing)
}

// Broadcast sends a message to all responders at this location. This
// implements the broadcast.Interface - see that for more details.
func (l *Location) Broadcast(omit []thing.Interface, format string, any ...interface{}) {
	msg := fmt.Sprintf("\n"+format, any...)

	for _, v := range l.Inventory.List(omit...) {
		if resp, ok := v.(responder.Interface); ok {
			resp.Respond(msg)
		}
	}
}

// Process implements the command.Interface to handle location specific
// commands. First we see if anything at the location can process the command
// and then the location itself. By handling commands in this order anything at
// a location: doors, barriers, guards, etc - can effect movement easily.
func (l *Location) Process(cmd *command.Command) (handled bool) {

	if handled = l.Inventory.Delegate(cmd); handled {
		return
	}

	switch cmd.Verb {
	case "LOOK", "L":
		handled = l.Look(cmd)
	case "EXITS", "EX":
		handled = l.exits(cmd)
	case "NORTH", "N":
		handled = l.move(cmd, NORTH)
	case "NORTHEAST", "NE":
		handled = l.move(cmd, NORTHEAST)
	case "EAST", "E":
		handled = l.move(cmd, EAST)
	case "SOUTHEAST", "SE":
		handled = l.move(cmd, SOUTHEAST)
	case "SOUTH", "S":
		handled = l.move(cmd, SOUTH)
	case "SOUTHWEST", "SW":
		handled = l.move(cmd, SOUTHWEST)
	case "WEST", "W":
		handled = l.move(cmd, WEST)
	case "NORTHWEST", "NW":
		handled = l.move(cmd, NORTHWEST)
	case "UP", "U":
		handled = l.move(cmd, UP)
	case "DOWN", "D":
		handled = l.move(cmd, DOWN)
	}

	return
}

// BUG(Diddymus): The Java version listed mobiles before other things in Look.

// Look implements the 'LOOK' command. It describes the location displaying the
// title, description, things and directional exits.
//
// TODO: Implement brief mode.
//
// TODO: Implement looking in a specific direction with a maximum viewing
// distance.
func (l *Location) Look(cmd *command.Command) (handled bool) {

	list := l.Inventory.List(cmd.Issuer)
	thingsHere := make([]string, 0, len(list))
	for _, o := range list {
		thingsHere = append(thingsHere, "You can see "+o.Name()+" here.")
	}

	things := ""
	if len(thingsHere) > 0 {
		things = strings.Join(thingsHere, "\n") + "\n"
	}

	cmd.Respond("[CYAN]%s[WHITE]\n%s\n[GREEN]%s\n[CYAN]You can see exits: [YELLOW]%s", l.Name(), l.Description(), things, l.directionalExits)

	return true
}

// exits implements the 'EXITS' command. It display the currently available
// directional exits from the location.
func (l *Location) exits(cmd *command.Command) (handled bool) {
	cmd.Respond("[CYAN]You can see exits: [YELLOW]%s", l.directionalExits)
	return true
}

// move implements the directional movement commands. This allows movement from
// location to location by typing a direction such as N or North.
//
// TODO: Modify command so that it can handle buffering of multiple location
// broadcasts.
func (l *Location) move(cmd *command.Command, d direction) (handled bool) {
	if to := l.directionalExits[d]; to != nil {
		if !cmd.CanLock(to) {
			cmd.AddLock(to)
			return true
		}

		l.Remove(cmd.Issuer)
		l.Broadcast([]thing.Interface{cmd.Issuer}, "[YELLOW]You see %s go %s.", cmd.Issuer.Name(), directionNames[d])

		to.Add(cmd.Issuer)
		to.Broadcast([]thing.Interface{cmd.Issuer}, "[YELLOW]You see %s walk in.", cmd.Issuer.Name())

		to.Look(cmd)
	} else {
		cmd.Respond("You can't go %s from here!", directionNames[d])
	}
	return true
}
