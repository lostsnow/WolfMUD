package entities

import (
	"fmt"
)

/*
	The direction type is defined so that functions can take a direction for the
	exits array index. Then a user can only pass one of the valid defined
	direction values.  Note that this type is not exported which would allow user
	defined and probably invalid values to be created.
*/
type direction uint8

/*
	Direction constants of type direction used to index exits array. Only these valid
	constants can be used. Note that the constants ARE exported while the type is
	not. This is valid as a user will refer to the constants and not the type.
*/
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

var directionNames = [10]string{
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

type Location interface {
	Thing
	Inventory
	Looker
	move(c Cmd, d direction) (handled bool)
	LinkExit(d direction, to Location)
}

type location struct {
	thing
	inventory
	exits [10]Location
}

func NewLocation(name, alias, description string) Location {
	return &location{
		thing:     thing{name, alias, description},
		inventory: inventory{},
	}
}

func (l *location) Command(c Cmd) (handled bool) {
	switch c.Verb() {
	case "LOOK":
		handled = l.look(c)
	case "NORTH", "N":
		handled = l.move(c, NORTH)
	case "NORTHEAST", "NE":
		handled = l.move(c, NORTHEAST)
	case "EAST", "E":
		handled = l.move(c, EAST)
	case "SOUTHEAST", "SE":
		handled = l.move(c, SOUTHEAST)
	case "SOUTH", "S":
		handled = l.move(c, SOUTH)
	case "SOUTHWEST", "SW":
		handled = l.move(c, SOUTHWEST)
	case "WEST", "W":
		handled = l.move(c, WEST)
	case "NORTHWEST", "NW":
		handled = l.move(c, NORTHWEST)
	case "UP":
		handled = l.move(c, UP)
	case "DOWN":
		handled = l.move(c, DOWN)
	}

	if handled == false {
		handled = l.thing.Command(c)
		if handled == false {
			handled = l.inventory.delegate(c)
		}
	}

	return handled
}

func (l *location) look(c Cmd) (handled bool) {
	if c.Target() != nil {
		return false
	}

	fmt.Printf("\n%s\n\n%s\n", l.name, l.description)

	for _, v := range l.inventory.List(c.What()) {
		fmt.Printf("You can see %s here\n", v.Name())
	}

	fmt.Println("")
	return true
}

func (l *location) LinkExit(d direction, to Location) {
	l.exits[d] = to
}

func (from *location) move(c Cmd, d direction) (handled bool) {
	if to := from.exits[d]; to != nil {
		from.Remove(c.What().Alias(), 1)
		to.Add(c.What())
		if m, ok := c.What().(Mobile); ok {
			m.Locate(to)
		}
		fmt.Printf("You go %s.\n", directionNames[d])
		to.look(c)
	} else {
		fmt.Printf("You can't go %s from here!\n", directionNames[d])
	}
	return true
}
