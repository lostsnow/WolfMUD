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
	Responder
	move(cmd Command, d direction) (handled bool)
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

func (l *location) Process(cmd Command) (handled bool) {
	switch cmd.Verb {
	case "LOOK":
		handled = l.look(cmd)
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
	case "UP":
		handled = l.move(cmd, UP)
	case "DOWN":
		handled = l.move(cmd, DOWN)
	}

	if handled == false {
		handled = l.thing.Process(cmd)
	}

	if handled == false {
		handled = l.inventory.delegate(cmd)
	}

	return handled
}

func (l *location) look(cmd Command) (handled bool) {
	if cmd.Target != nil {
		return false
	}

	msg := fmt.Sprintf("\n%s\n\n%s\n", l.name, l.description)

	for _, v := range l.inventory.List(cmd.Issuer) {
		msg += fmt.Sprintf("You can see %s here\n", v.Name())
	}

	cmd.Respond(msg)

	return true
}

func (l *location) LinkExit(d direction, to Location) {
	l.exits[d] = to
}

func (from *location) move(cmd Command, d direction) (handled bool) {
	if to := from.exits[d]; to != nil {
		from.Remove(cmd.Issuer.Alias(), 1)
		from.Respond("You see %s go %s.\n", cmd.Issuer.Name(), directionNames[d])

		if m, ok := cmd.Issuer.(Mobile); ok {
			m.Locate(to)
		}

		cmd.Respond("You go %s.\n", directionNames[d])
		to.Respond("You see %s walk in.\n", cmd.Issuer.Name())
		to.Add(cmd.Issuer)

		to.look(cmd)
	} else {
		cmd.Respond("You can't go %s from here!\n", directionNames[d])
	}
	return true
}

func (l *location) Respond(format string, any ...interface{}) {
	msg := fmt.Sprintf(format, any...)
	for _, v := range l.inventory.List(nil) {
		if resp, ok := v.(Responder); ok {
			resp.Respond(msg)
		}
	}
}
