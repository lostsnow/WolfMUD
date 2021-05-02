// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"strconv"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for holdable attribute.
func init() {
	internal.AddMarshaler((*Holdable)(nil), "holdable")
}

// Holdable implements an attribute for specifying body slots required when
// holding a Thing. Holdable will veto the JUNK command so that items being
// held are not accidentally junked and disposed of.
//
// Unlike Wieldable and Wearable types any Thing that is not a player or mobile,
// does not have a Body attribute, is always Holdable in one hand.
//
// This can be overridden by adding a specific Holdable attribute - for example
// if two hands are required to hold the Thing:
//
//  Holdable: HAND→2
//
// Or if you want to be able to hold a small animal in one hand:
//
//  Holdable: HAND
//
// If an item should not be holdable at all the HOLD command can be vetoed.
type Holdable struct {
	Attribute
	slots []string
}

// Some interfaces we want to make sure we implement
var (
	_ has.Holdable = &Holdable{}
	_ has.Vetoes   = &Holdable{}
	_ has.Slotable = &Holdable{}
)

// NewHoldable returns a new Holdable attribute initialised with the passed
// Body slot references. Any Thing with a Holdable attribute can be held by a
// player or mobile provided they have the specified Body slots available.
func NewHoldable(slots ...string) *Holdable {
	return &Holdable{Attribute{}, slots}
}

// FindHoldable searches the attributes of the specified Thing for attributes
// that implement has.Holdable returning the first match it finds.
//
// If not match is found and the Thing is not a player or mobile, does not have
// a Body attribute, then a default Holdable with a single 'HAND' slot will be
// returned.
//
// If the Thing is a player or mobile, has a Body attribute, then a *Holdable
// typed nil will be returned.
func FindHoldable(t has.Thing) has.Holdable {
	h := t.FindAttr((*Holdable)(nil)).(has.Holdable)
	if !h.Found() && !FindBody(t).Found() {
		h = NewHoldable("HAND")
		h.SetParent(t)
	}
	return h
}

// Is returns true if passed attribute implements a holdable item else false.
func (*Holdable) Is(a has.Attribute) bool {
	_, ok := a.(has.Holdable)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (h *Holdable) Found() bool {
	return h != nil
}

// Unmarshal is used to turn the passed data into a new Holdable attribute.
func (*Holdable) Unmarshal(data []byte) has.Attribute {

	slots := []string{}

	for slot, count := range decode.PairList(data) {
		c := 1
		if len(count) > 0 {
			if i, err := strconv.Atoi(count); err == nil {
				c = i
			}
		}
		for x := 0; x < c; x++ {
			slots = append(slots, slot)
		}
	}
	return NewHoldable(slots...)
}

// Marshal returns a tag and []byte that represents the receiver.
func (h *Holdable) Marshal() (tag string, data []byte) {

	iSlots := make(map[string]int, len(h.slots))    // Slots + integer quantities
	sSlots := make(map[string]string, len(h.slots)) // Slots + string quantities

	for _, slot := range h.slots {
		iSlots[slot]++
	}

	for slot, count := range iSlots {
		sSlots[slot] = strconv.Itoa(count)
	}

	return "holdable", encode.PairList(sSlots, '→')
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (h *Holdable) Dump(node *tree.Node) *tree.Node {
	slots := []byte{}
	if len(h.slots) > 0 {
		for _, slot := range h.slots {
			slots = strconv.AppendQuote(append(slots, ", "...), slot)
		}
		slots = slots[2:]
	}
	return node.Append("%p %[1]T - slots: %d [%s]", h, len(h.slots), slots)
}

// IsHoldable returns true.
func (h *Holdable) IsHoldable() bool {
	return true
}

// Slots returns the Body slot references that need to be available to hold
// the Thing. The returned slice should not be modified.
func (h *Holdable) Slots() []string {
	return h.slots
}

// Check will veto the JUNK command if the Thing is currently held.
func (h *Holdable) Check(actor has.Thing, cmd ...string) has.Veto {

	for _, cmd := range cmd {

		// If command not JUNK we won't veto
		if cmd != "JUNK" {
			continue
		}

		// If actor has no body it can't be holding so we won't veto
		b := FindBody(actor)
		if !b.Found() {
			continue
		}

		if usage := b.Usage(h.Parent()); usage != "" {
			name := FindName(h.Parent()).TheName("something")
			return NewVeto(cmd, "You connot junk "+name+" while "+usage+" it.")
		}
	}
	return nil
}

// Copy returns a copy of the Holdable receiver.
func (h *Holdable) Copy() has.Attribute {
	if h == nil {
		return (*Holdable)(nil)
	}
	return NewHoldable(h.Slots()...)
}
