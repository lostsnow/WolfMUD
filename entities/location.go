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
	Looker
	Move(what Thing, d direction) (handled bool)
	LinkExit(d direction, to Location)
	Add(t Thing)
	Delete(t Thing)
}

type location struct {
	thing
	exits    [10]Location
	contains []Thing
}

func NewLocation(name, description string, l Location) Location {
	return &location{
		thing: thing{name, description, l},
	}
}

func (l *location) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	default:
		handled = l.thing.Command(what, cmd, args)
	case "LOOK":
		handled = l.Look(what, args)
	case "NORTH", "N":
		handled = l.Move(what, NORTH)
	case "EAST", "E":
		handled = l.Move(what, EAST)
	case "SOUTH", "S":
		handled = l.Move(what, SOUTH)
	case "WEST", "W":
		handled = l.Move(what, WEST)
	}
	return handled
}

func (l *location) Look(what Thing, args []string) (handled bool) {
	if len(args) == 0 {
		fmt.Printf("\n%s\n\n%s\n\n", l.name, l.description)
		if len(l.contains) > 0 {
			msg := ""
			for _, t := range l.contains {
				if t != what {
					msg += fmt.Sprintf("\t%s\n", t.Name())
				}
			}
			if msg != "" {
				fmt.Printf("You can see here:\n%s\n", msg)
			}
		}
	}
	return true
}

func (l *location) LinkExit(d direction, to Location) {
	l.exits[d] = to
}

func (l *location) Add(t Thing) {
	l.contains = append(l.contains, t)
}

func (l *location) Delete(t Thing) {
	for k, v := range l.contains {
		if v == t {
			l.contains = append(l.contains[:k], l.contains[k+1:]...)
			break
		}
	}
}

func (from *location) Move(what Thing, d direction) (handled bool) {
	if to := from.exits[d]; to != nil {
		from.Delete(what)
		to.Add(what)
		fmt.Printf("You go %s.\n", directionNames[d])
		to.Look(what, nil)
		what.Locate(to)
	} else {
		fmt.Printf("You can't go %s from here!\n", directionNames[d])
		to = from
	}
	return true
}
