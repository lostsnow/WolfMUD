// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package proc

import (
	"strings"
)

type isAttr byte
type asAttr byte

// Is Attributes
const (
	Unset isAttr = 1 << iota
	Start
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
	North asAttr = iota
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

// Direction mappings
var (
	NameToDir = map[string]asAttr{
		"N": North, "NE": Northeast, "E": East, "SE": Southeast,
		"S": South, "SW": Southwest, "W": West, "NW": Northwest,
		"NORTH": North, "NORTHEAST": Northeast, "EAST": East, "SOUTHEAST": Southeast,
		"SOUTH": South, "SOUTHWEST": Southwest, "WEST": West, "NORTHWEST": Northwest,
		"UP": Up, "DOWN": Down,
	}
	DirToName = map[asAttr]string{
		North: "north", Northeast: "northeast", East: "east", Southeast: "southeast",
		South: "south", Southwest: "southwest", West: "west", Northwest: "northwest",
		Up: "up", Down: "down",
	}
)

// Thing is a basic one thing fits all type.
type Thing struct {
	Name        string
	Description string
	Is          isAttr
	As          map[asAttr]string
	In          []*Thing
}

func NewThing(name, description string) *Thing {
	return &Thing{
		Name:        name,
		Description: description,
		As:          make(map[asAttr]string),
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
