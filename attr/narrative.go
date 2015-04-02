// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Narrative implements an attribute for non-removable content. It allows
// creators to cater to the more discerning adventurer by providing content
// that is not spoon fed to them. Narrative content is usually mentioned or
// discoverable from text descriptions. For example:
//
//	You are in the corner of a common room in the Dragon's Breath tavern. There
//	is a fire burning away merrily in an ornate fireplace giving comfort to
//	weary travellers. Shadows flicker around the room, changing light to
//	darkness and back again. To the south the common room extends and east the
//	common room leads to the tavern entrance.
//
// From such a description it would be reasonable for someone to want to
// example the fireplace although there would be no "You see a fireplace here."
// when listing the items at the location. Should someone try to examine the
// fireplace they are rewarded with:
//
//	This is a very ornate fireplace carved from marble. Either side a dragon
//	curls downward until the head is below the fire looking upward, giving the
//	impression that they are breathing fire.
//
// While anything that can normally be put into an inventory can be put into a
// narrative, nothing should be directly removable. However everything in a
// narrative still works as expected - readable things are still readable and
// containers can have things put in them as well as removed. As an example
// consider this brief description:
//
//	You are standing next to a small fish pond. Paths lead off north, south and
//	west deeper into the gardens.
//
// Examining the pond - in this case a simple inventory - reveals its content:
//
//	This is a small fish pond. It contains a fish of gold.
//
// Taking the fish from the pond and examining it reveals:
//
//	This is a small fish made from solid gold.
//
// A much more satisfying reward for being curious :)
//
// NOTE: At the moment narrative content should not be removeable for the
// simple reason that descriptions are mostly static - for now(?). So removing
// something would therefore invalidate the descriptive text.
//
// BUG(diddymus): Currently narrative Things should implement a veto on GET to
// prevent removal. Narratives should prevent removal automatically - perhaps
// by reimplementing the Inventory Remove method? But then how do we remove a
// narrative if we really, really, we know what we are doing, want to? For
// example we could want to remove a Thing and add a similar Thing to, in
// effect, change something. Maybe allow access to (simple conversion?) the
// embeded Inventory directly for removals?
type Narrative struct {
	*Inventory
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
var (
	_ has.Inventory = &Narrative{}
	_ has.Narrative = &Narrative{}
)

// NewNarrative returns a new Narrative attribute initialised with the
// specified Things as initial contents.
func NewNarrative(t ...has.Thing) *Narrative {
	return &Narrative{NewInventory(t...)}
}

// FindNarrative searches the attributes of the specified Thing for attributes
// that implement has.Narrative returning the first match it finds or nil
// otherwise.
func FindNarrative(t has.Thing) has.Narrative {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Narrative); ok {
			return a
		}
	}
	return nil
}

// ImplementsNarrative is a marker method so that we can distinguish between
// a Narrative (that is also an Inventory) and a plain Inventory.
func (n *Narrative) ImplementsNarrative() {}

func (n *Narrative) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T %d items:", n, len(n.contents)))
	for _, n := range n.contents {
		for _, i := range n.Dump() {
			buff = append(buff, DumpFmt("%s", i))
		}
	}
	return buff
}
