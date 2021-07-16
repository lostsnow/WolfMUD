// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"strings"
)

// Type definitions for Thing field keys.
type (
	isKey  uint32 // index for Thing.Is
	asKey  uint32 // index for Thing.As
	anyKey uint32 // index for Thing.Any
	intKey uint32 // index for Thing.Int
)

// Constants for use as bitmasks with the Thing.Is field.
const (
	Container isKey = 1 << iota // A container, allows PUT/TAKE
	Dark                        // A dark location
	Location                    // Item is a location
	NPC                         // An NPC
	Narrative                   // A narrative item
	Open                        // An open item (e.g. door)
	Player                      // Is a player
	Start                       // A starting location
)

// isNames maps isKey bits to their string name. See also setName method.
var isNames = []string{
	"Container",
	"Dark",
	"Location",
	"NPC",
	"Narrative",
	"Open",
	"Player",
	"Start",
}

// setNames returns the names of the set bits in a Thing.Is field. Names are
// separated by the OR (|) symbol. For example: "Narrative|Open".
func (is isKey) setNames() string {
	names := []string{}
	for x := len(isNames) - 1; x >= 0; x-- {
		if is&(1<<x) != 0 {
			names = append(names, isNames[x])
		}
	}
	return strings.Join(names, "|")
}

// Constants for use as keys in a Thing.As field.
//
// NOTE: The first 10 direction constants are fixed and their values SHOULD NOT
// BE CHANGED. The other constants should be kept in alphabetical order as new
// ones are added.
const (

	// Location reference exit leads to ("L1")
	North asKey = iota
	Northeast
	East
	Southeast
	South
	Southwest
	West
	Northwest
	Up
	Down

	Blocker          // Name of direction being blocked ("E")
	Description      // Item's description
	DynamicAlias     // "PLAYER" or unset, "SELF" for actor performing a command
	DynamicQualifier // Situation dependant e.g. GET sets "MY",DROP deleted "MY"
	Name             // Item's name
	OnCleanup        // Custome cleanup message for an item
	OnReset          // Custom reset message for an item
	Ref              // Item's original reference (zone:ref or ref)
	UID              // Item's unique identifier
	VetoDrop         // Veto for DROP command
	VetoGet          // Veto for GET command
	VetoPut          // Veto PUT command for item
	VetoPutIn        // Veto for PUT command into container
	VetoTake         // Veto TAKE command for item
	VetoTakeOut      // Veto for TAKE command from container
	Where            // Current location ref ("L1")
	Writing          // Description of writing on an item
	Zone             // Zone item's definition loaded from
)

// asNames maps asKey values to their string name.
var asNames = []string{
	"North", "Northeast", "East", "Southeast",
	"South", "Southwest", "West", "Northwest",
	"Up", "Down",

	"Blocker",
	"Description",
	"DynamicAlias",
	"DynamicQualifier",
	"Name",
	"OnCleanup",
	"OnReset",
	"Reference",
	"UID",
	"VetoDrop",
	"VetoGet",
	"VetoPut",
	"VetoPutIn",
	"VetoTake",
	"VetoTakeOut",
	"Where",
	"Writing",
	"Zone",
}

var (
	// NameToDir maps a long or short direction name to its Thing.As constant.
	NameToDir = map[string]asKey{
		"N": North, "NE": Northeast, "E": East, "SE": Southeast,
		"S": South, "SW": Southwest, "W": West, "NW": Northwest,
		"U": Up, "D": Down,
		"NORTH": North, "NORTHEAST": Northeast, "EAST": East, "SOUTHEAST": Southeast,
		"SOUTH": South, "SOUTHWEST": Southwest, "WEST": West, "NORTHWEST": Northwest,
		"UP": Up, "DOWN": Down,
	}

	// DirToName maps a Thing.As direction constant to the direction's long name.
	DirToName = map[asKey]string{
		North: "north", Northeast: "northeast", East: "east", Southeast: "southeast",
		South: "south", Southwest: "southwest", West: "west", Northwest: "northwest",
		Up: "up", Down: "down",
	}
)

// ReverseDir returns the reverse or opposite direction. For example if passed
// the constant East it will return West. If the passed value is not one of the
// direction constants it will be returned unchanged.
func (dir asKey) ReverseDir() asKey {
	switch {
	case dir > Down:
		return dir
	case dir < Up:
		return dir ^ 1<<2
	default:
		return dir ^ 1
	}
}

// Constants for Thing.Any keys
const (
	Alias     anyKey = iota // Aliases for an item
	OnAction                // Actions that can be performed
	Qualifier               // Alias qualifiers
)

// anyNames maps anyKey values to their string name.
var anyNames = []string{
	"Alias",
	"OnAction",
	"Qualifier",
}

// Constants for Thing.Int keys
const (
	ActionAfter   intKey = iota // How often an action event should occur
	ActionJitter                // Maximum random delay to add to ActionAfter
	CleanupAfter                // How soon a clean-up event should occur
	CleanupJitter               // Maximum random delay to add to CleanupAfter
	ResetAfter                  // How soon a reset event should occur
	ResetJitter                 // Maximum random delay to add to TesetAfter
)

// intNames maps intKey values to their string name.
var intNames = []string{
	"ActionAfter",
	"ActionJitter",
	"CleanupAfter",
	"CleanupJitter",
	"ResetAfter",
	"ResetJitter",
}