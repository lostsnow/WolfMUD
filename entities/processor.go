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
	Processor is an interface implemented by entities that have command
	processing capabilities. Commands are typically what a user types. For
	example:

		GET BALL

	However the command interface provides a very loose coupling that entities
	can make use of as well. For example if you had a lever in a wall a player
	might try:

		PULL LEVER

	To which the lever could issue the following to a door entity:

		OPEN DOOR

	Net result is player pulls lever and a door opens. However the door entity
	should check that only a lever entity is issuing the command and it is not
	from a player!

	In effect this could provide a way of dynamically scripting entity behaviour.
*/
type Processor interface {
	Process(Command) (handled bool)
}

/*
	Command is a reference to an unexported command struct. It is returned via a
	call to NewCommand. As command structs are passed around alot between
	entities Command makes sure we are passing command structs by reference.
*/
type Command *command

/*
	command is a structure that eases handling of command strings. When a new
	command is created it contains:

		Issuer - What is issuing the command, usually a mobile/player

		Verb - The verb in the command. E.G. GET, DROP, LOOK, etc...

		Nouns - The nouns the verb should act on. E.G. BALL, SWORD, GOLD, etc...

		Target - A shortcut to command.Nouns[0] - the first noun. Most commands
		only act on a single noun: GET BALL, DROP SWORD
*/
type command struct {
	Issuer  Thing
	Verb    string
	Nouns   []string
	Target  *string
	Respond	func(format string, any ...interface{})
}

/*
	NewCommand takes a thing that wants to execute a command, the command as a
	string and returns a new command struct wrapped as a Command type. This can
	then be passed to types implementing the Processor interface to do fun and
	useful things.

	See command struct for more detail.
*/
func NewCommand(issuer Thing, input string) Command {
	words := strings.Split(strings.ToUpper(input), ` `)

	cmd := command{}

	cmd.Issuer = issuer
	cmd.Verb = words[0]
	cmd.Nouns = words[1:]
	cmd.Respond = func(format string, any ...interface{}) {
		if resp, ok := cmd.Issuer.(Responder); ok {
			resp.Respond(format, any...)
		}
	}

	if len(words) > 1 {
		cmd.Target = &words[1]
	}

	return &cmd
}
