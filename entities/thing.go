/*
	Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.
*/

package entities

import (
	"fmt"
	"strings"
)

/*
	Thing is a type representing the most basic type of entity. It is the
	lowest denominator from which most other entities are built.

	As the thing struct is not exported the Thing type defines accessor methods
	for retrieving some of a thing's fields.
*/
type Thing interface {
	Examiner
	Commander
	Name() (name string)
	Alias() (name string)
}

type thing struct {
	name        string   // The name of a thing
	alias       string   // An alias to refer to a thing
	description string   // A description of the thing
}

/*
	NewThing is a constructor to create things of type Thing. A thing cannot be
	created directly because it is not exported, however the Thing type is
	exported and acts to provide external access.
*/
func NewThing(name, alias, description string) Thing {
	return &thing{name, strings.ToUpper(alias), description}
}

/*
	Name returns the name for a thing.
*/
func (t *thing) Name() string {
	return t.name
}

/*
	Alias returns the alias command processing uses to identify a thing. For example:

		> GET BALL

	This would cause the command processing to look for a thing with an alias of
	'BALL' and act on it using the 'GET' command.
*/
func (t *thing) Alias() string {
	return t.alias
}

func (t *thing) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	case "LOOK":
		handled = t.look(what, args)
	case "EXAMINE":
		handled = t.examine(what, args)
	}
	return handled
}

func (t *thing) look(what Thing, args []string) (handled bool) {
	if len(args) == 0 {
		return false
	}

	if args[0] == t.Alias() {
		fmt.Printf("You look at %s. %s\n", t.name, t.description)
		return true
	}
	return false
}

func (t *thing) examine(what Thing, args []string) (handled bool) {
	if len(args) == 0 {
		return false
	}

	if args[0] == t.Alias() {
		fmt.Printf("You examine %s. %s\n", t.name, t.description)
		return true
	}
	return false
}
