// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

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
)

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

var directionIndex = map[string]byte{
	"N":         North,
	"NORTH":     North,
	"NE":        Northeast,
	"NORTHEAST": Northeast,
	"E":         East,
	"EAST":      East,
	"SE":        Southeast,
	"SOUTHEAST": Southeast,
	"S":         South,
	"SOUTH":     South,
	"SW":        Southwest,
	"SOUTHWEST": Southwest,
	"W":         West,
	"WEST":      West,
	"NW":        Northwest,
	"NORTHWEST": Northwest,
	"U":         Up,
	"UP":        Up,
	"D":         Down,
	"DOWN":      Down,
}

type Exits struct {
	Attribute
	exits [len(directionNames)]has.Thing
}

// Some interfaces we want to make sure we implement
var (
	_ has.Exits = &Exits{}
)

func NewExits() *Exits {
	return &Exits{Attribute{}, [len(directionNames)]has.Thing{}}
}

func FindExits(t has.Thing) has.Exits {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Exits); ok {
			return a
		}
	}
	return nil
}

func (e *Exits) Dump() []string {
	buff := []byte{}
	for i, e := range e.exits {
		if e != nil {
			buff = append(buff, ", "...)
			buff = append(buff, directionNames[i]...)
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

func Return(direction byte) byte {
	if direction < Up {
		return direction ^ 1<<2
	}
	return direction ^ 1
}

func (e *Exits) Link(direction byte, to has.Thing) {
	e.exits[direction] = to
}

func (e *Exits) AutoLink(direction byte, to has.Thing) {
	e.Link(direction, to)
	if E := FindExits(to); E != nil {
		E.Link(Return(direction), e.Parent())
	}
}

func (e *Exits) Unlink(direction byte) {
	e.exits[direction] = nil
}

func (e *Exits) AutoUnlink(direction byte) {
	to := e.exits[direction]
	e.Unlink(direction)

	if to == nil {
		return
	}

	if E := FindExits(to); E != nil {
		E.Unlink(Return(direction))
	}
}

func (e *Exits) List() string {

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

func (e *Exits) Place(t has.Thing) {
	if a := FindInventory(e.Parent()); a != nil {
		a.Add(t)
	}
}

func (e *Exits) Move(t has.Thing, cmd string) (msg string, ok bool) {

	d, valid := directionIndex[cmd]

	if !valid {
		msg = "You wanted to go which way!?"
		return
	}

	if e.exits[d] == nil {
		msg = "You can't go " + directionNames[d] + " from here!"
		return
	}

	from := FindInventory(e.Parent())
	if from == nil {
		msg = "You are not sure where you are, let alone where you are going."
		return
	}

	to := FindInventory(e.exits[d])
	if to == nil {
		msg = "For some odd reason you can't go " + directionNames[d] + "."
		return
	}

	if what := from.Remove(t); what == nil {
		msg = "Something stops you from leaving here!"
		return
	}

	to.Add(t)

	return "", true
}
