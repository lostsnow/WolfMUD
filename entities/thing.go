package entities

import (
	"fmt"
	"strings"
)

type Looker interface {
	Look(what Thing, args []string) (handled bool)
}

type Examiner interface {
	Examine(what Thing, args []string) (handled bool)
}

type Commander interface {
	Command(what Thing, cmd string, args []string) (handled bool)
}

// A basic Thing

type Thing interface {
	Examiner
	Commander
	Name() (name string)
	Alias() (name string)
	Locate(l Location)
}

type thing struct {
	name        string
	alias       string
	description string
	location    Location
}

func NewThing(name, alias, description string, location Location) Thing {
	return &thing{name, strings.ToUpper(alias), description, location}
}

func (t *thing) Name() string {
	return t.name
}

func (t *thing) Alias() string {
	return t.alias
}

func (t *thing) Locate(l Location) {
	t.location = l
}

func (t *thing) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	default:
		handled = false // thing has nothing further to delegate to :(
	case "LOOK":
		handled = t.Look(what, args)
	case "EXAMINE":
		handled = t.Examine(what, args)
	}
	return handled
}

func (t *thing) Look(what Thing, args []string) (handled bool) {
	if len(args) == 0 {
		return false
	}

	if args[0] == t.Alias() {
		fmt.Printf("You look at %s. %s\n", t.name, t.description)
		return true
	}
	return false
}

func (t *thing) Examine(what Thing, args []string) (handled bool) {
	if len(args) == 0 {
		return false
	}

	if args[0] == t.Alias() {
		fmt.Printf("You examine %s. %s\n", t.name, t.description)
		return true
	}
	return false
}
