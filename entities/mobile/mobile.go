// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package mobile defines the most basic type of mobile. A mobile is any
// computer controlled non-player character and includes creatures, monsters,
// fogs, gelatinous cubes or anything else you can think of.
package mobile

import (
	"log"
	"code.wolfmud.org/WolfMUD.git/entities/location"
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/utils/command"
	"code.wolfmud.org/WolfMUD.git/utils/inventory"
)

// Mobile provides a default basic implementation of a mobile.
type Mobile struct {
	thing.Thing
	inventory.Inventory
	location location.Interface
}

// New creates a new Mobile and returns a reference to it.
func New(name string, alias []string, description string) *Mobile {
	m := &Mobile{
		Thing:     *thing.New(name, alias, description),
		Inventory: *inventory.New(),
	}

	log.Printf("Mobile %d created: %s\n", m.UniqueId(), m.Name())

	return m
}

// Relocate sets a mobile's internal location reference. It implements part of
// the location.Locateable interface.
func (m *Mobile) Relocate(l location.Interface) {
	m.location = l
}

// Locate gets a mobile's internal location reference. It implements part of
// the location.Locateable interface.
func (m *Mobile) Locate() location.Interface {
	return m.location
}

// Process implements the command.Interface to handle location specific
// commands.
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

// inv implements the 'INVENTORY' command. This provides information about what
// a mobile is currently carrying. Currently a mobile can only examine it's own
// inventory - but someone like a theif might find it handy to look into someone
// elses inventory ;)
//
// TODO: Currently very basic, needs to deal with held, weilded, worn items.
func (m *Mobile) inv(cmd *command.Command) (handled bool) {

	if cmd.Target != "" {
		return
	}

	response := ""

	if inventory := m.Inventory.List(); len(inventory) == 0 {
		response = "You are not carrying anything."
	} else {
		response = "You are currently carrying:\n"
		for _, item := range inventory {
			response += "\t" + item.Name() + "\n"
		}
	}
	cmd.Respond(response)

	return true
}
