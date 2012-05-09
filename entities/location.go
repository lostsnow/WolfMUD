package entities

import (
	"fmt"
	"strings"
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
		thing:     *NewThing(name, alias, description).(*thing),
		inventory: *NewInventory().(*inventory),
	}
}

func (l *location) Process(cmd Command) (handled bool) {
	switch cmd.Verb {
	case "LOOK", "L":
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

	availableExits := []string{}

	for direction, location := range l.exits {
		if location != nil {
			availableExits = append(availableExits, directionNames[direction])
		}
	}

	availableThings := []string{}

	for _, v := range l.inventory.List(cmd.Issuer) {
		availableThings = append(availableThings, "You can see "+v.Name()+" here")
	}
	if len(availableThings) > 0 {
		availableThings = append(availableThings, "")
	}

	msg := fmt.Sprintf("%s\n%s\n%s\nYou can see exits: %s", l.name, l.description, strings.Join(availableThings, "\n"), strings.Join(availableExits, ", "))

	cmd.Respond(msg)

	return true
}

func (l *location) LinkExit(d direction, to Location) {
	l.exits[d] = to
}

func (from *location) move(cmd Command, d direction) (handled bool) {
	if to := from.exits[d]; to != nil {
		from.RespondGroup([]Thing{cmd.Issuer}, "You see %s go %s.", cmd.Issuer.Name(), directionNames[d])
		from.Remove(cmd.Issuer.Alias(), 1)

		cmd.Respond("You go %s.", directionNames[d])
		to.Add(cmd.Issuer)
		to.RespondGroup([]Thing{cmd.Issuer}, "You see %s walk in.", cmd.Issuer.Name())

		to.look(cmd)
	} else {
		cmd.Respond("You can't go %s from here!", directionNames[d])
	}
	return true
}

func (l *location) RespondGroup(ommit []Thing, format string, any ...interface{}) {
	msg := fmt.Sprintf(format, any...)

OMMIT:
	for _, v := range l.inventory.List(nil) {
		if resp, ok := v.(Responder); ok {
			for _, o := range ommit {
				if o.IsAlso(v) {
					continue OMMIT
				}
			}
			resp.Respond(msg)
		}
	}
}

func (l *location) Respond(format string, any ...interface{}) {
	msg := fmt.Sprintf(format, any...)
	for _, v := range l.inventory.List(nil) {
		if resp, ok := v.(Responder); ok {
			resp.Respond(msg)
		}
	}
}

func (l *location) Add(t Thing) {
	l.inventory.Add(t)
	t.Locate(l)
}

func (l *location) Remove(alias string, occurance int) (t Thing) {
	t = l.inventory.Remove(alias, occurance)
	t.Locate(nil)
	return
}
