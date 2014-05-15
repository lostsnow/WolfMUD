// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package item implements a very basic, general item that can be picked up and
// carried around.
package item

import (
	"code.wolfmud.org/WolfMUD.git/entities/location"
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/utils/command"
	"code.wolfmud.org/WolfMUD.git/utils/inventory"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"
	"code.wolfmud.org/WolfMUD.git/utils/units"

	"log"
)

// Item type is a default implementation of an item.
type Item struct {
	thing.Thing
	weight units.Weight
}

// Register zero value instance of Item with the loader.
func init() {
	recordjar.RegisterUnmarshaler("item", &Item{})
}

func (i *Item) Unmarshal(d recordjar.Decoder) {
	i.weight = units.Weight(d.Int("weight"))
	i.Thing.Unmarshal(d)
}

// TODO: Instead of calling Unmarshal within Init we should be calling a
// Copy/Clone function instead.
func (i *Item) Init(d recordjar.Decoder, refs map[string]recordjar.Unmarshaler) {
	for x, location := range d.KeywordList("location") {
		if l, ok := refs[location]; ok {
			if l, ok := l.(inventory.Interface); ok {
				if x == 0 {
					l.Add(i)
				} else {
					tmp := &Item{}
					tmp.Unmarshal(d)
					l.Add(tmp)
				}
				log.Printf("Added %s to %s", i.Name(), location)
			} else {
				log.Printf("Cannot add %q to %q: Not an inventory", i.Name(), location)
			}
		} else {
			log.Printf("Cannot add %q to %q: Ref not found", i.Name(), location)
		}
	}
}

// Process implements the command.Interface to handle Item specific
// commands.
func (i *Item) Process(cmd *command.Command) (handled bool) {

	// This specific item?
	if !i.IsAlias(cmd.Target) {
		return
	}

	switch cmd.Verb {
	case "DROP":
		handled = i.drop(cmd)
	case "WEIGH":
		handled = i.weigh(cmd)
	case "EXAMINE", "EXAM":
		handled = i.examine(cmd)
	case "GET":
		handled = i.get(cmd)
	case "JUNK":
		handled = i.junk(cmd)
	}

	return
}

// drop removes an Item from the command issuer's inventory and puts it into
// the inventory of the issuer's current location. For this to happen a few
// conditions must be true:
//
//	1. Issuer must be at some sort of location
//	2. Issuer must implement an inventory
//	3. Inventory must contain the requested item
//
func (i *Item) drop(cmd *command.Command) (handled bool) {
	if m, ok := cmd.Issuer.(location.Locateable); ok {
		if inv, ok := cmd.Issuer.(inventory.Interface); ok {
			if inv.Contains(i) {
				inv.Remove(i)
				cmd.Respond("You drop %s.", i.Name())
				cmd.Broadcast([]thing.Interface{cmd.Issuer}, "You see %s drop %s.", cmd.Issuer.Name(), i.Name())

				m.Locate().Add(i)

				handled = true
			}
		}
	} else {
		cmd.Respond("You don't see anywhere to drop %s.", i.Name())
		cmd.Broadcast([]thing.Interface{cmd.Issuer}, "You see %s try and drop %s.", cmd.Issuer.Name(), i.Name())
		handled = true
	}
	return
}

// weigh estimates the weight of the specified item.
func (i *Item) weigh(cmd *command.Command) (handled bool) {
	cmd.Respond("You estimate %s to weigh about %s.", i.Name(), i.weight)
	cmd.Broadcast([]thing.Interface{cmd.Issuer}, "You see %s estimate the weight of %s.", cmd.Issuer.Name(), i.Name())
	return true
}

// examine describes the specific item.
func (i *Item) examine(cmd *command.Command) (handled bool) {
	cmd.Respond("You examine %s. %s", i.Name(), i.Description())
	cmd.Broadcast([]thing.Interface{cmd.Issuer}, "You see %s study %s.", cmd.Issuer.Name(), i.Name())
	return true
}

// get removes an Item from the command issuer's current location and puts it
// into it's own inventory. For this to happen a few conditions must be true:
//
//	1. Issuer must be at some sort of location
//	2. Issuer must implement an inventory
//	3. Issuer's location must contain the requested item
//
func (i *Item) get(cmd *command.Command) (handled bool) {
	if m, ok := cmd.Issuer.(location.Locateable); ok {
		if inv, ok := cmd.Issuer.(inventory.Interface); ok {
			if l := m.Locate(); l.Contains(i) {
				l.Remove(i)
				cmd.Broadcast([]thing.Interface{cmd.Issuer}, "You see %s pick up %s.", cmd.Issuer.Name(), i.Name())

				inv.Add(i)
				cmd.Respond("You pickup %s.", i.Name())

				handled = true
			}
		}
	}
	return
}

// TODO: Implement junk command
func (i *Item) junk(cmd *command.Command) (handled bool) {
	cmd.Respond("Junk not implemented yet.")
	return true
}
