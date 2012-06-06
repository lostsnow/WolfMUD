/*
	Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.

	Use of this source code is governed by the license in the LICENSE file
	included with the source code.
*/

package mobile

import (
	"log"
	"runtime"
	"wolfmud.org/entities/location"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
	"wolfmud.org/utils/inventory"
)

type Interface interface {
	Relocate(l location.Interface)
}

type Mobile struct {
	*thing.Thing
	*inventory.Inventory
	location location.Interface
}

func New(name string, alias []string, description string) *Mobile {
	m := &Mobile{
		Thing:     thing.New(name, alias, description),
		Inventory: inventory.New(),
	}

	log.Printf("Mobile %d created: %s\n", m.UniqueId(), m.Name())
	runtime.SetFinalizer(m, final)

	return m
}

func final(m *Mobile) {
	log.Printf("+++ Mobile %d finalized: %s +++\n", m.UniqueId(), m.Name())
}

func (m *Mobile) Relocate(l location.Interface) {
	m.location = l
}

func (m *Mobile) Locate() location.Interface {
	return m.location
}

func (m *Mobile) Process(cmd *command.Command) (handled bool) {

	switch cmd.Verb {
	case "INVENTORY", "INV":
		handled = m.inv(cmd)
	}

	if m.IsAlso(cmd.Issuer) {
		if handled == false {
			handled = m.Inventory.Delegate(cmd)
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
