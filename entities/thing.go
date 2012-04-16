package entities

import (
	"fmt"
)

type Looker interface {
	Look(what Thing, args []string) (handled bool)
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
