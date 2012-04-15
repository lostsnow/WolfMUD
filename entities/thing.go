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
	Look(what Thing)
}

type Examiner interface {
	Examine()
}

type Commander interface {
	Command(what Thing, cmd string, args []string) (handled bool)
}

// A basic Thing

type Thing interface {
	Examiner
	Commander
	Name() (name string)
	Locate(l Location)
}

type thing struct {
	name        string
	description string
	location    Location
}

func NewThing(name, description string, location Location) Thing {
	return &thing{
		name,
		description,
		location,
	}
}

func (t *thing) Name() string {
	return t.name
}

func (t *thing) Locate(l Location) {
	t.location = l
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

func NewDroper(name, description string, location Location) Dropper {
	return &droper{
		thing{
			name,
			description,
			location,
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
	Looker
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

func NewLocation(name, description string, l Location) Location {
	return &location{
		thing: thing{
			name,
			description,
			l,
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
		l.Look(what)
	case "NORTH", "N":
		l.Move(what, NORTH)
	case "EAST", "E":
		l.Move(what, EAST)
	case "SOUTH", "S":
		l.Move(what, SOUTH)
	case "WEST", "W":
		l.Move(what, WEST)
	}
	return true
}

func (l *location) Look(what Thing) {
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

func (from *location) Move(what Thing, d direction) (to Location) {
	if to = from.exits[d]; to != nil {
		from.Delete(what)
		to.Add(what)
		fmt.Printf("You go %s.\n", directionNames[d])
		to.Look(what)
		what.Locate(to)
	} else {
		fmt.Printf("You can't go %s from here!\n", directionNames[d])
		to = from
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

func NewPlayer(name, description string, location Location) (p Player) {
	return &player{
		thing: thing{
			name,
			description,
			location,
		},
	}
}

func (p *player) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	default:
		if handled = p.location.Command(what, cmd, args); handled == true {
		} else if handled = p.thing.Command(what, cmd, args); handled == true {
		}
		return handled
	}
	return true
}
