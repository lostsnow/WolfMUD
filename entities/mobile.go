/*
	Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.

	Use of this source code is governed by the license in the LICENSE file
	included with the source code.
*/

package entities

import (
	"fmt"
)

/*
	Mobile is an interface representing the most basic type of 'living' thing. It
	is the lowest denominator from which most other creatures and players are
	built.

	The Mobile interface embeds Thing and Inventory - A Mobile must be able to
	describe itself and needs to be able to carry things.

	As the mobile struct is not exported the Mobile type defines accessor methods
	for retrieving some of a thing's fields.
*/
type Mobile interface {
	Thing
	Inventory
	Parse(cmd string)
}

/*
	The mobile type embeds thing providing basic functionality such as name and
	description. It also embeds inventory so that any mobile type can carry
	things around. These satisfy the Thing and Inventory interfaces embeded in
	the Mobile interface.
*/
type mobile struct {
	thing
	inventory
}

func NewMobile(name, alias, description string) Mobile {
	return &mobile{
		thing: *NewThing(name, alias, description).(*thing),
	}
}

func (m *mobile) Parse(input string) {
	fmt.Printf("\n> %s\n", input)
	handled := m.Process(NewCommand(m, input))
	if handled == false {
		fmt.Printf("Eh? %s?\n\n", input)
	}
}

/*
	Process satisfies the Processor interface and implements the main processing
	for commands by mobiles. This is also the main starting point for commands
	from players. Commands are handled or delegated to:

		1. The current Mobile - INVENTORY, SCORE

		2. Current Mobile's inventory - DROP BALL, EXAMINE BALL

		3. The Mobile's environment/current location - LOOK, NORTH, N

		4. Things at the current location - GET BALL, KILL DIDDYMUS

	Items 2-4 are only processed if the mobile is also the mobile issuing the
	command - i.e. the mobile is itself.
*/
func (m *mobile) Process(cmd Command) (handled bool) {

	switch cmd.Verb {
	case "INVENTORY", "INV":
		handled = m.inv(cmd)
	}

	// Pass up to embeded thing?
	if handled == false {
		handled = m.thing.Process(cmd)
	}

	if m.IsAlso(cmd.Issuer) {
		if handled == false {
			handled = m.inventory.delegate(cmd)
		}

		if handled == false {
			l := m.location
			if: l != nil {
				handled = l.Process(cmd)
			}
		}
	}

	return
}

/*
  inv lists the things a mobile is carrying in it's inventory.

	Currently we only handle inventory for the current mobile. If a target has
	been specified (i.e. someone else) we don't process the command but bail out
	early. In future we may be able to show the inventory for others - be
	interesting for some theiving skills perhaps?
*/
func (m *mobile) inv(cmd Command) (handled bool) {

	response := ""

	if cmd.Target != nil {
		return false
	} else {
		if inventory := m.inventory.List(cmd.Issuer); len(inventory) == 0 {
			response = "You are not carrying anything.\n"
		} else {
			response = "You are currently carrying:\n"
			for _, item := range inventory {
				response += "\t" + item.Name() + "\n"
			}
		}
	}
	cmd.Respond(response)

	return true
}
