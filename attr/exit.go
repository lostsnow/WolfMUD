// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/has"

	"errors"
	"strings"
)

// Constants for direction indexes. These can be used for the Link, AutoLink,
// Unlink and AutoUnlink methods. If these constants are modified probably need
// to update the Return function as well.
const (
	North byte = iota
	Northeast
	East
	Southeast
	South
	Southwest
	West
	Northwest
	Up
	Down

	exitCount // Make sure this is the last item!
)

// directionNames is a lookup table for direction indexes to direction strings.
// When listing available exits they will be presented in the order they are in
// in this array.
var directionNames = [...]string{
	North:     "north",
	Northeast: "northeast",
	East:      "east",
	Southeast: "southeast",
	South:     "south",
	Southwest: "southwest",
	West:      "west",
	Northwest: "northwest",
	Up:        "up",
	Down:      "down",
}

// directionIndex is a lookup table for direction strings to direction indexes.
// The directional strings cover upper, lower and title cased directions. See
// also NormalizeDirection method.
var directionIndex = map[string]byte{
	"N":         North,
	"n":         North,
	"NORTH":     North,
	"north":     North,
	"North":     North,
	"NE":        Northeast,
	"ne":        Northeast,
	"NORTHEAST": Northeast,
	"northeast": Northeast,
	"Northeast": Northeast,
	"E":         East,
	"e":         East,
	"EAST":      East,
	"east":      East,
	"East":      East,
	"SE":        Southeast,
	"se":        Southeast,
	"SOUTHEAST": Southeast,
	"southeast": Southeast,
	"Southeast": Southeast,
	"S":         South,
	"s":         South,
	"SOUTH":     South,
	"south":     South,
	"South":     South,
	"SW":        Southwest,
	"sw":        Southwest,
	"SOUTHWEST": Southwest,
	"southwest": Southwest,
	"Southwest": Southwest,
	"W":         West,
	"w":         West,
	"WEST":      West,
	"west":      West,
	"West":      West,
	"NW":        Northwest,
	"nw":        Northwest,
	"NORTHWEST": Northwest,
	"northwest": Northwest,
	"Northwest": Northwest,
	"U":         Up,
	"u":         Up,
	"UP":        Up,
	"up":        Up,
	"Up":        Up,
	"D":         Down,
	"d":         Down,
	"DOWN":      Down,
	"down":      Down,
	"Down":      Down,
}

// Exits implements an attribute describing exits for the eight compass points
// north, northeast, east, southeast, south, southwest, west and northwest as
// well as the directions up and down and where they lead to. Exits are usually
// in pairs, for example one north and one back south. You can have one way
// exits or return exits that do not lead back to where you came from.
type Exits struct {
	Attribute
	exits [exitCount]has.Inventory
}

// Some interfaces we want to make sure we implement
var (
	_ has.Exits = &Exits{}
)

// NewExits returns a new Exits attribute with no exits set. Exits should be
// added to the attribute using the Link and AutoLink methods. The reason exits
// cannot be set during initialisation like most other attributes is that all
// 'locations' have to be setup before they can all be linked together.
func NewExits() *Exits {
	return &Exits{Attribute{}, [exitCount]has.Inventory{}}
}

// FindExits searches the attributes of the specified Thing for attributes that
// implement has.Exits returning the first match it finds or a *Exits typed nil
// otherwise.
func FindExits(t has.Thing) has.Exits {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Exits); ok {
			return a
		}
	}
	return (*Exits)(nil)
}

// Found returns false if the receiver is nil otherwise true. This is a utility
// method that can be chained with FindExits to easily check if an Exits
// attribute was found.
func (e *Exits) Found() bool {
	return e != nil
}

func (e *Exits) Dump() []string {
	buff := []byte{}
	for i, e := range e.exits {
		if e != nil {
			buff = append(buff, ", "...)
			buff = append(buff, directionNames[i]...)
			buff = append(buff, ": "...)
			buff = append(buff, FindName(e.Parent()).Name("Somewhere")...)
		}
	}
	if len(buff) > 0 {
		buff = buff[2:]
	}
	return []string{DumpFmt("%p %[1]T -> %s", e, buff)}
}

// Return calculates the opposite/return direction for the direction given.
// This is handy for calculating things like normal exits where if you go north
// you return by going back south. It is also useful for implementing ranged
// weapons, thrown weapons and spells. For example if you fire a bow west the
// person will see the arrow come from the east (from their perspective).
func Return(direction byte) byte {
	if direction < Up {
		return direction ^ 1<<2
	}
	return direction ^ 1
}

// Link links the given exit direction to the given Inventory. If the given
// direction was already linked the exit will be overwritten - in effect the
// same as unlinking the exit first and then relinking it.
func (e *Exits) Link(direction byte, to has.Inventory) {
	if e != nil {
		e.exits[direction] = to
	}
}

// AutoLink links the given exit, calculates the opposite return exit and links
// that automatically as well - as long as the parent Thing of the to Inventory
// has an Exits attribute.
func (e *Exits) AutoLink(direction byte, to has.Inventory) {
	e.Link(direction, to)
	FindExits(to.Parent()).Link(Return(direction), FindInventory(e.Parent()))
}

// Unlink sets the exit for the given direction to nil. It does not matter if
// the given direction was not linked in the first place.
func (e *Exits) Unlink(direction byte) {
	e.Link(direction, nil)
}

// AutoUnlink unlinks the given exit, calculates the opposite return exit and
// unlinks that automatically as well.
//
// BUG(diddymus): Does not check that exit A links to B and B links back to A.
// For example a maze may have an exit going North from A to B but going South
// from B takes you to C instead of back to A as would be expected!
func (e *Exits) AutoUnlink(direction byte) {
	if e == nil {
		return
	}

	e.Unlink(direction)
	if to := e.exits[direction]; to != nil {
		FindExits(to.Parent()).Unlink(Return(direction))
	}
}

// List will return a string listing the exits you can see. For example:
//
//	You can see exits east, southeast and south.
//
func (e *Exits) List() string {

	if e == nil {
		return "You can see no immediate exits from here."
	}

	// Note we can tell the difference between l=0 initially and l=0 when the
	// last location was North by looking at the count c. If c is zero we have
	// not found any exits. If c is not zero then l=0 represents North.
	var (
		buff = make([]byte, 0, 1024) // buffer for direction list
		l    = 0                     // direction index of last exit found
		c    = 0                     // count of useable (linked) exits found
	)

	for i, e := range e.exits {
		switch {
		case e == nil:
			continue
		case c > 1:
			buff = append(buff, ", "...)
			fallthrough
		case c > 0:
			buff = append(buff, directionNames[l]...)
		}
		c++
		l = i
	}

	switch c {
	case 0:
		return "You can see no immediate exits from here."
	case 1:
		return "The only exit you can see from here is " + directionNames[l] + "."
	default:
		return "You can see exits " + string(buff) + " and " + directionNames[l] + "."
	}
}

// NormalizeDirection takes a long or short variant of a direction name in any
// case and returns the direction.
//
// So 'N', 'NORTH', 'n', 'north', 'North' and 'NoRtH' all return the constant
// NORTH which is 0.
//
// If the direction name given cannot be normalized, maybe because it is
// invalid, a non-nil error will be returned.
func (_ *Exits) NormalizeDirection(name string) (direction byte, err error) {

	// Common case quick path - upper, lower or title cased input
	if d, valid := directionIndex[name]; valid {
		return d, nil
	}

	// Slower path - force name to all uppercase and trim any possible whitespace
	name = strings.ToUpper(strings.TrimSpace(name))

	// See if we can find a match now
	if d, valid := directionIndex[name]; valid {
		return d, nil
	}

	// No matches found so assume direction name is invalid
	return 0, errors.New("invalid direction")
}

// ToName returns the lowercased long name of a direction or an empty string if
// the direction is invalid.
func (_ *Exits) ToName(direction byte) (name string) {
	if direction >= 0 && direction < exitCount {
		return directionNames[direction]
	}
	return ""
}

// LeadsTo returns the Inventory of the location found by taking a specific
// exit. If a particular direction leads nowhere nil will be returned.
func (e *Exits) LeadsTo(direction byte) has.Inventory {
	if e != nil {
		return e.exits[direction]
	}
	return nil
}

// Within returns all of the locations within the given number of moves from
// the location specified by the receiver. It is 3D and will follow up and down
// exits as well. The locations are returned as a slice of Inventory slices.
// The first slice represents the cwnumber of moves. The second slice is all
// locations within that number of moves. The first slice, moves 0, is always
// the current location.
//
// Assume the following map of the Tavern and surrounding locations:
//
//                ____________________________________________________________
//               |1             |3             |5              |6             |
//               |  Fireplace       Entrance     Between           Bakery     |
//               |                               Tavern/Bakery                |
//               |__                         __|__           __|______________|
//               |2              4             |7              |8             |
//               |  Common Room     Bar        | Street outside    Pawn Shop  |
//               |                             | Pawn Shop                    |
//               |______________|______________|__           __|______________|
//                              |10            |9              |
//                              |   Outside      Fountain      |
//                              |   Armourer     Square        |
//                              |______________|_______________|
//
//
// If we are at location 7 on the map then Within(2) will return all locations
// within 2 moves:
//
//	 [][]has.Inventory{
//		[]has.Inventory{ 7 },					// Within 0 moves of location 7
//		[]has.Inventory{ 5, 8, 9 },   // Within 1 move of location 7
//		[]has.Inventory{ 6, 3, 10 },  // Within 2 moves of location 7
//	 }
//
// The above numbers e.g. 5,8,9 refer to the map locations. In reality they
// would actually be references to has.Inventory interface types.
//
// See cmd/sneeze.go for an example of using the Within method.
func (e *Exits) Within(moves int) (locations [][]has.Inventory) {
	if e == nil {
		return
	}

	// Add current location at zero moves
	center := FindInventory(e.Parent())
	locations = append(locations, []has.Inventory{center})

	// Go through each number of moves expanding from the center/current location
	for m := 1; m <= moves; m++ {

		// M is the group of locations for a given number of moves m
		M := []has.Inventory{}

		// Go through locations found for previous number of moves to see if there
		// are more exits available to follow
		for _, e := range locations[m-1] {
		exitLoop:
			for _, l := range FindExits(e.Parent()).Surrounding() {
				// For location l go through each number of moves mm for each location
				// MM in the results we have so far to see if l is in the results yet
				for _, mm := range locations {
					for _, MM := range mm {
						if MM == l {
							continue exitLoop
						}
					}
				}
				// If location not found add to group M (locations at m moves out)
				M = append(M, l)
			}
		}

		// Add group M (locations at m moves out) to results
		locations = append(locations, M)
	}

	return
}

// Surrounding returns an Inventory slice of all locations immediatly reachable
// from the current location. The cuurent location is specified by the
// receiver. If there are no immediatly reachable locations an empty slice will
// be returned.
func (e *Exits) Surrounding() []has.Inventory {

	exits := make([]has.Inventory, 0, exitCount)

	if e == nil {
		return exits
	}

	for _, i := range e.exits {
		if i != nil {
			exits = append(exits, i)
		}
	}
	return exits
}
