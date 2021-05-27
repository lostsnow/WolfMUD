// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package proc

import (
	"strings"
)

// Is Attributes
const (
	Start uint32 = 1 << iota
	Narrative
	Dark
	NPC
)

// Is value mapping to name.
var isNames = []string{
	"Start", "Narrative", "Dark", "NPC",
}

// isNames returns the names of the set flags separated by the OR (|) symbol.
func IsNames(is uint32) string {
	names := []string{}
	for x := len(isNames) - 1; x >= 0; x-- {
		if is&(1<<x) != 0 {
			names = append(names, isNames[x])
		}
	}
	return strings.Join(names, "|")
}

// As value keys
const (
	North uint32 = iota
	Northeast
	East
	Southeast
	South
	Southwest
	West
	Northwest
	Up
	Down
	Where
	Alias
)

// As value mappings
var asNames = []string{
	"N", "NE", "E", "SE", "S", "SW", "W", "NW", "Up", "Down",
	"Where", "Alias",
}

// Direction mappings
var (
	// NameToDir maps a long or short direction name to its As constant.
	NameToDir = map[string]uint32{
		"N": North, "NE": Northeast, "E": East, "SE": Southeast,
		"S": South, "SW": Southwest, "W": West, "NW": Northwest,
		"NORTH": North, "NORTHEAST": Northeast, "EAST": East, "SOUTHEAST": Southeast,
		"SOUTH": South, "SOUTHWEST": Southwest, "WEST": West, "NORTHWEST": Northwest,
		"UP": Up, "DOWN": Down,
	}

	// DirToName maps an As direction constant to the direction's long name.
	DirToName = map[uint32]string{
		North: "north", Northeast: "northeast", East: "east", Southeast: "southeast",
		South: "south", Southwest: "southwest", West: "west", Northwest: "northwest",
		Up: "up", Down: "down",
	}
)

var nextUID chan uint32

func init() {
	nextUID = make(chan uint32, 1)
	nextUID <- 0
}

// Thing is a basic one thing fits all type.
type Thing struct {
	Name        string
	Description string
	UID         uint32
	Is          uint32
	As          map[uint32]string
	In          []*Thing
}

func NewThing(name, description string) *Thing {
	uid := <-nextUID
	nextUID <- uid + 1
	return &Thing{
		UID:         uid,
		Name:        name,
		Description: description,
		As:          make(map[uint32]string),
	}
}

// Find looks for a Thing with the given alias in the provided list of Things
// inventories. If a matching Thing is found returns the Thing, the Thing who's
// Inventory it was in and the index in the inventory where it was found. If
// there is not match returns nill for the Thing, nil for the Inventory and an
// index of -1.
func Find(alias string, where ...*Thing) (*Thing, *Thing, int) {
	if alias == "" {
		return nil, nil, -1
	}
	for _, inv := range where {
		if inv == nil {
			continue
		}
		for idx, item := range inv.In {
			if item.As[Alias] == alias {
				return item, inv, idx
			}
		}
	}
	return nil, nil, -1
}
