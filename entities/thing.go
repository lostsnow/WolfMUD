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

type Looker interface {
	Look()
}

type Examiner interface {
	Examine()
}

type Commander interface {
	Command(what Thing, cmd string, args []string) (handled bool)
}

// A basic Thing

type Thing interface {
	Looker
	Examiner
	Commander
	Name() (name string)
}

type thing struct {
	name        string
	description string
}

func NewThing(name, description string) Thing {
	return &thing{
		name,
		description,
	}
}

func (t *thing) Name() string {
	return t.name
}

func (t *thing) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	default:
		return false
	case "LOOK":
		t.Look()
	case "EXAMINE":
		t.Examine()
	}
	return true
}

func (t *thing) Look() {
	fmt.Printf("You look at %s\n%s\n", t.name, t.description)
}

func (t *thing) Examine() {
	fmt.Printf("You examine %s\n%s\n", t.name, t.description)
}

// A droppable Thing

type Dropper interface {
	Thing
	Drop()
}

type droper struct {
	thing
}

func NewDroper(name, description string) Dropper {
	return &droper{
		thing{
			name,
			description,
		},
	}
}

func (d *droper) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	default:
		return d.thing.Command(what, cmd, args)
	case "DROP":
		d.Drop()
	}
	return true
}

func (d *droper) Drop() {
	fmt.Printf("You drop %s\n", d.name)
	return
}

// A location

type Location interface {
	Thing
	Move(what Thing, d direction) (to Location)
	SetExit(d direction, to Location)
	Add(t Thing)
	Delete(t Thing)
}

type location struct {
	thing
	exits    [10]Location
	contains []Thing
}

func NewLocation(name, description string) Location {
	return &location{
		thing: thing{
			name,
			description,
		},
	}
}

func (l *location) SetExit(d direction, to Location) {
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

func (l *location) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	default:
		return l.thing.Command(what, cmd, args)
	case "LOOK":
		l.Look()
	case "NORTH", "N":
		l.Move(what, NORTH)
	case "EAST", "E":
		l.Move(what, EAST)
	case "SOUTH", "S":
		l.Move(what, SOUTH)
	}
	return true
}

func (l *location) Look() {
	fmt.Printf("\n%s\n\n%s\n\n", l.name, l.description)
	if len(l.contains) > 0 {
		fmt.Printf("You can see here:\n")
		for _, t := range l.contains {
			fmt.Printf("\t%s\n", t.Name())
		}
	}
}

func (l *location) Move(what Thing, d direction) (to Location) {
	if to = l.exits[d]; to != nil {
		l.Delete(what)
		fmt.Printf("You go %s.\n", directionNames[d])
		to.Look()
		to.Add(what)
	} else {
		fmt.Printf("You can't go %s from here!\n", directionNames[d])
		to = l
	}
	return
}

// A player

type Player interface {
	Thing
}

type player struct {
	thing
}

func NewPlayer(name, description string) Player {
	return &player{
		thing: thing{
			name,
			description,
		},
	}
}

func (l *player) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	default:
		return l.thing.Command(what, cmd, args)
	}
	return true
}
