// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package inventory implements a 'collection' of Things. It can be used as the
// inventory of a location - who and what is at that location. It can also be
// used as a literal inventory for container type objects such as bags, boxes
// and pouches.
package inventory

import (
	"wolfmud.org/entities/thing"
)

// Interface describes the methods required to be an Inventory.
type Interface interface {
	Add(thing thing.Interface)
	Remove(thing thing.Interface)
	List(ommit ...thing.Interface) []thing.Interface
}

// Inventory is a default implementation satisfiying the inventory interface.
type Inventory struct {
	contents []thing.Interface
}

// New creates a new Inventory and returns a pointer reference to it.
func New() *Inventory {
	return &Inventory{}
}

// Add puts an object implementing thing.Interface into the Inventory.
func (i *Inventory) Add(thing thing.Interface) {
	i.contents = append(i.contents, thing)
}

// Remove takes an object implementing thing.Interface from the Inventory.
func (i *Inventory) Remove(thing thing.Interface) {
	for index, t := range i.contents {
		if t.IsAlso(thing) {
			i.contents = append(i.contents[:index], i.contents[index+1:]...)
			break
		}
	}
}

// List returns a slice of thing.Interface in the Inventory, possibly with
// specific items ommited. An example of when you want to ommit something is
// when a Player does something - you send a specific message to the player:
//
//  You pick up a ball.
//
// A different message is sent to any observers:
//
//  You see Diddymus pick up a ball.
//
// However when broadcasting the message to the location you want to ommit the
// 'actor' who has the specific message.
//
// Note that locations implement an inventory to store what mobiles/players and
// things are present which is why this works.
func (i *Inventory) List(ommit ...thing.Interface) (list []thing.Interface) {

OMMIT:
	for _, thing := range i.contents {
		for _, o := range ommit {
			if thing.IsAlso(o) {
				continue OMMIT
			}
			list = append(list, thing)
		}
	}

	return
}
