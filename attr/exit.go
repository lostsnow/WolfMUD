// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

const (
	N, NORTH uint8 = iota, iota
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

var directionLongNames = [...]string{
	N:  "north",
	NE: "northeast",
	E:  "east",
	SE: "southeast",
	S:  "south",
	SW: "southwest",
	W:  "west",
	NW: "northwest",
	U:  "up",
	D:  "down",
}

var minLongNamesSize = 0

var directionIndex = map[string]uint8{
	"N":         N,
	"NORTH":     N,
	"NE":        NE,
	"NORTHEAST": NE,
	"E":         E,
	"EAST":      E,
	"SE":        SE,
	"SOUTHEAST": SE,
	"S":         S,
	"SOUTH":     S,
	"SW":        SW,
	"SOUTHWEST": SW,
	"W":         W,
	"WEST":      W,
	"NW":        NW,
	"NORTHWEST": NW,
	"U":         U,
	"UP":        U,
	"D":         D,
	"DOWN":      D,
}

func init() {
	minLongNamesSize = 0
	for _, d := range directionLongNames {
		minLongNamesSize += len(d) + 1
	}
}

type exits struct {
	parent
	exits [len(directionLongNames)]has.Thing
}

func NewExits() *exits {
	return &exits{parent{}, [len(directionLongNames)]has.Thing{}}
}

func FindExit(t has.Thing) has.Exit {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Exit); ok {
			return a
		}
	}
	return nil
}

func (e *exits) Dump() []string {
	buff := []byte{}
	for i, e := range e.exits {
		if e != nil {
			buff = append(buff, ", "...)
			buff = append(buff, directionLongNames[i]...)
			buff = append(buff, ": "...)
			if a := FindName(e); a != nil {
				buff = append(buff, a.Name()...)
			}
		}
	}
	if len(buff) > 0 {
		buff = buff[2:]
	}
	return []string{DumpFmt("%p %[1]T -> %s", e, buff)}
}

func (e *exits) Link(direction uint8, to has.Thing) {
	e.exits[direction] = to
}

func (e *exits) Unlink(direction uint8) {
	e.exits[direction] = nil
}

func (e *exits) Description() string {
	buff := make([]byte, 0, minLongNamesSize)

	for i, e := range e.exits {
		if e != nil {
			buff = append(buff, directionLongNames[i]...)
			buff = append(buff, ' ')
		}
	}

	if len(buff) == 0 {
		return "You can see no immediate exits from here."
	}
	return "You can see exits: " + string(buff)
}

func (e *exits) Place(t has.Thing) string {
	if a := FindInventory(e.Parent()); a != nil {
		a.Add(t)
	}
	return ""
}

func (e *exits) Move(t has.Thing, cmd string) string {

	// Check exit available
	d := directionIndex[cmd]
	if e.exits[d] == nil {
		return "You can't go " + directionLongNames[d] + " from here!"
	}

	// Remove mover from current exit's parent inventory
	if a := FindInventory(e.Parent()); a != nil {
		a.Remove(t)
	}

	// Add mover to new exit's parent inventory
	// NOTE: The exit already points to the parent so we don't need to find it.
	if a := FindInventory(e.exits[d]); a != nil {
		a.Add(t)
	}

	return ""
}
