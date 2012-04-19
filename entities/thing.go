/*
	Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.

	Use of this source code is governed by the license in the LICENSE file
	included with the source code.
*/

package entities

import (
	"strings"
)

/*
	Thing is an interface representing the most basic type of entity. It is the
	lowest denominator from which most other entities are built.

	As the thing struct is not exported the Thing type defines accessor methods
	for retrieving some of a thing's fields.
*/
type Thing interface {
	Examiner
	Processor
	Name() (name string)
	Alias() (name string)
}

type thing struct {
	name        string // The name of a thing
	alias       string // An alias to refer to a thing
	description string // A description of the thing
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
	Alias returns the alias that command processing uses to identify a thing. For
	example:

		> GET BALL

	This would cause the command processing to look for a thing with an alias of
	'BALL' and act on it using the 'GET' command.
*/
func (t *thing) Alias() string {
	return t.alias
}

/*
	Process satisfies the Processor interface and implements the main processing
	for commands usable on a Thing.
*/
func (t *thing) Process(cmd Command) (handled bool) {

	// Is the command for 'this' thing?
	if cmd.Target == nil || *cmd.Target != t.alias {
		return
	}

	switch cmd.Verb {
	case "LOOK", "L":
		handled = t.look(cmd)
	case "EXAMINE", "EX":
		handled = t.examine(cmd)
	}

	return
}

/*
	look processes the 'Look' or 'L' command for a Thing.
*/
func (t *thing) look(cmd Command) (handled bool) {
	cmd.Respond("You look at %s. %s\n", t.name, t.description)
	return true
}

/*
	examine processes the 'Examine' or 'Ex' command for a Thing.
*/
func (t *thing) examine(cmd Command) (handled bool) {
	cmd.Respond("You examine %s. %s\n", t.name, t.description)
	return true
}
