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
	Move(what Thing, d direction) (handled bool)
	LinkExit(d direction, to Location)
}

type location struct {
	thing
	inventory
	exits [10]Location
}

func NewLocation(name, alias, description string, l Location) Location {
	return &location{
		thing:     thing{name, alias, description, l},
		inventory: inventory{},
	}
}

func (l *location) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	case "LOOK":
		handled = l.Look(what, args)
	case "NORTH", "N":
		handled = l.Move(what, NORTH)
	case "NORTHEAST", "NE":
		handled = l.Move(what, NORTHEAST)
	case "EAST", "E":
		handled = l.Move(what, EAST)
	case "SOUTHEAST", "SE":
		handled = l.Move(what, SOUTHEAST)
	case "SOUTH", "S":
		handled = l.Move(what, SOUTH)
	case "SOUTHWEST", "SW":
		handled = l.Move(what, SOUTHWEST)
	case "WEST", "W":
		handled = l.Move(what, WEST)
	case "NORTHWEST", "NW":
		handled = l.Move(what, NORTHWEST)
	case "UP":
		handled = l.Move(what, UP)
	case "DOWN":
		handled = l.Move(what, DOWN)
	}

	if handled == false {
		handled = l.thing.Command(what, cmd, args)
		if handled == false {
			handled = l.inventory.delegate(what, cmd, args)
		}
	}

	return handled
}

func (l *location) Look(what Thing, args []string) (handled bool) {
	if len(args) != 0 {
		return false
	}

	fmt.Printf("\n%s\n\n%s\n", l.name, l.description)

	if list := l.List(what); list != "" {
		fmt.Printf(l.List(what))
	}
	fmt.Println("")
	return true
}

func (l *location) LinkExit(d direction, to Location) {
	l.exits[d] = to
}

func (from *location) Move(what Thing, d direction) (handled bool) {
	if to := from.exits[d]; to != nil {
		from.Remove(what.Alias(), 1)
		to.Add(what)
		what.Locate(to)
		fmt.Printf("You go %s.\n", directionNames[d])
		to.Look(what, nil)
	} else {
		fmt.Printf("You can't go %s from here!\n", directionNames[d])
	}
	return true
}
