/*
	Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.

	Use of this source code is governed by the license in the LICENSE file
	included with the source code.
*/

package mobile

import (
	"wolfmud.org/entities/inventory"
	"wolfmud.org/entities/location"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
)

type Interface interface {
	Locate(l location.Interface)
}

type Mobile struct {
	*thing.Thing
	*inventory.Inventory
	location location.Interface
}

func New(name string, alias []string, description string) *Mobile {
	return &Mobile{
		Thing:     thing.New(name, alias, description),
		Inventory: inventory.New(),
	}
}

func (m *Mobile) Locate(l location.Interface) {
	m.location = l
}

func (m *Mobile) Process(cmd *command.Command) (handled bool) {

	switch cmd.Verb {
	case "INVENTORY", "INV":
		handled = m.inv(cmd)
	}

	// Pass up to embeded thing?
	//if handled == false {
	//	handled = m.Thing.Process(cmd)
	//}

	if m.IsAlso(cmd.Issuer) {
		if handled == false {
			//handled = m.Inventory.delegate(cmd)
		}

		if handled == false {
			l := m.location
			if l != nil {
				handled = l.Process(cmd)
			}
		}
	}

	return
}

func (m *Mobile) inv(cmd *command.Command) (handled bool) {

	response := ""

	if cmd.Target != nil {
		return false
	} else {
		if inventory := m.Inventory.List(); len(inventory) == 0 {
			response = "You are not carrying anything."
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
