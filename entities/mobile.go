/*
	Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.

	Use of this source code is governed by the license in the LICENSE file
	included with the source code.
*/

package entities

import (
	"fmt"
)

type Mobile interface {
	Thing
	Inventory
	Parse(cmd string)
	Locate(l Location)
}

type mobile struct {
	thing
	inventory
	location Location
}

func NewMobile(name, alias, description string) (m Mobile) {
	return &mobile{
		thing: thing{name, alias, description},
	}
}

func (m *mobile) Parse(input string) {
	fmt.Printf("\n> %s\n", input)
	handled := m.Process(NewCommand(m, input))
	if handled == false {
		fmt.Printf("Eh? %s?\n\n", input)
	}
}

func (m *mobile) Locate(l Location) {
	m.location = l
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
	default:

		// Pass up to embeded thing? Still for mobile
		handled = m.thing.Process(cmd)

		// Pass up to embeded inventory? 'Self' only
		if handled == false && cmd.What == m {
			handled = m.inventory.delegate(cmd)
		}

		// Pass to current location? 'Self' only
		if handled == false && cmd.What == m {
			handled = m.location.Process(cmd)
		}

	case "INVENTORY", "INV":
		handled = m.inv(cmd)
	}

	return
}

/*
 */
func (m *mobile) inv(cmd Command) (handled bool) {

	// Currently we only handle inventory for the current mobile. If a target has
	// been specified (i.e. someone else) we can't process the command so bail
	// early. In future we may be able to show the inventory for others - be
	// interesting for some theiving skills perhaps?
	if cmd.Target != nil {
		return
	}

	response := "You are currently carrying:\n"
	for _, v := range m.inventory.List(cmd.What) {
		response += fmt.Sprintf("\t%s\n", v.Name())
	}
	cmd.Respond(response)

	return true
}
