// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Narrative attribute.
func init() {
	internal.AddMarshaler((*Narrative)(nil), "narrative")
}

// Narrative implements an attribute to mark non-removable content. It allows
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
type Narrative struct {
	Attribute
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
var (
	_ has.Narrative = &Narrative{}
)

// NewNarrative returns a new Narrative attribute.
func NewNarrative() *Narrative {
	return &Narrative{Attribute{}}
}

// FindNarrative searches the attributes of the specified Thing for attributes
// that implement has.Narrative returning the first match it finds or a
// *Narrative typed nil otherwise.
func FindNarrative(t has.Thing) has.Narrative {
	return t.FindAttr((*Narrative)(nil)).(has.Narrative)
}

// Is returns true if passed attribute implements a narrative else false.
func (*Narrative) Is(a has.Attribute) bool {
	_, ok := a.(has.Narrative)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (n *Narrative) Found() bool {
	return n != nil
}

// Unmarshal is used to turn the passed data into a new Narrative attribute.
func (*Narrative) Unmarshal(data []byte) has.Attribute {
	return NewNarrative()
}

// Marshal returns a tag and []byte that represents the receiver.
func (n *Narrative) Marshal() (tag string, data []byte) {
	return "narrative", data
}

// ImplementsNarrative is a marker method so that we can specifically identify
// a Narrative.
func (*Narrative) ImplementsNarrative() {}

// Dump adds attribute information to the passed tree.Node for debugging.
func (n *Narrative) Dump(node *tree.Node) *tree.Node {
	return node.Append("%p %[1]T", n)
}

// Copy returns a copy of the Narrative receiver.
func (n *Narrative) Copy() has.Attribute {
	if n == nil {
		return (*Narrative)(nil)
	}
	return NewNarrative()
}
