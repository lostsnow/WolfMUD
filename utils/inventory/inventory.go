// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package inventory implements a 'collection' of Things. It can be used as the
// inventory of a location - who and what is at that location. It can also be
// used as a literal inventory for container type objects such as bags, boxes
// and pouches.
//
// TODO: implement inventory capacity
package inventory

import (
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
)

const (
	NOT_FOUND = -1 // Used if Thing not found in Inventory
)

// Interface describes the methods required to be an Inventory.
type Interface interface {
	Add(thing thing.Interface)
	Contains(thing thing.Interface) bool
	Remove(thing thing.Interface)
	List(omit ...thing.Interface) []thing.Interface
}

// Inventory is a default implementation satisfying the inventory interface.
type Inventory struct {
	contents []thing.Interface
}

// New creates a new Inventory and returns a pointer reference to it.
func New() *Inventory {
	return &Inventory{}
}

// Add puts an object implementing thing.Interface into the Inventory.
func (i *Inventory) Add(thing thing.Interface) {
	if i.find(thing) == NOT_FOUND {
		i.contents = append(i.contents, thing)
	}
}

// Contains returns true if an object implementing thing.Interface is in the
// Inventory, otherwise it returns false.
func (i *Inventory) Contains(thing thing.Interface) bool {
	return i.find(thing) != NOT_FOUND
}

// find is an internal helper which returns the index of a Thing in the
// Inventory. If the item is not in the Inventory then NOT_FOUND is returned.
// Ideally we do not want external users manipulating the indexes, therefore
// this method is not exported and users of the Inventory can use Contains to
// test for an object.
//
// Currently we brute force using a for/range and bail early when a match is
// found.
func (i *Inventory) find(thing thing.Interface) (index int) {
	for index, t := range i.contents {
		if t.IsAlso(thing) {
			return index
		}
	}
	return NOT_FOUND
}

// Remove takes an object implementing thing.Interface from the Inventory. If
// the inventory is now empty we trim the contents slice to set the length and
// capacity to zero to reclaim a little storage.
func (i *Inventory) Remove(thing thing.Interface) {
	if index := i.find(thing); index != NOT_FOUND {
		i.contents = append(i.contents[:index], i.contents[index+1:]...)
		if len(i.contents) == 0 {
			i.contents = nil
		}
	}
}

// List returns a slice of thing.Interface in the Inventory, possibly with
// specific items omitted. An example of when you want to omit something is when
// a Player does something - you send a specific message to the player:
//
//  You pick up a ball.
//
// A different message is sent to any observers:
//
//  You see Diddymus pick up a ball.
//
// However when broadcasting the message to the location you want to omit the
// 'actor' who has the specific message.
//
// Note that locations implement an inventory to store what mobiles/players and
// things are present which is why this works.
func (i *Inventory) List(omit ...thing.Interface) (list []thing.Interface) {

OMIT:
	for _, thing := range i.contents {
		for i, o := range omit {
			if thing.IsAlso(o) {
				omit = append(omit[0:i], omit[i+1:]...)
				continue OMIT
			}
		}
		list = append(list, thing)
	}

	return
}

// Delegate delegates commands to an invenotry's items. This is most useful
// when processing commands for a location and the location cannot process the
// command it passes it on to somethning else that might be able to.
func (i *Inventory) Delegate(cmd *command.Command) (handled bool) {
	for _, thing := range i.contents {

		// Don't process the command issuer - gets very recursive!
		if thing.IsAlso(cmd.Issuer) {
			continue
		}

		if thing, ok := thing.(command.Interface); ok {
			handled = thing.Process(cmd)
			if handled {
				return true
			}
		}
	}
	return false
}
