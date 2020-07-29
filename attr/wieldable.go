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

// Register marshaler for Wieldable attribute.
func init() {
	internal.AddMarshaler((*Wieldable)(nil), "wieldable")
}

// Wieldable implements an attribute for specifying body slots required when
// wielding a Thing. Wieldable will veto the JUNK command so that items being
// wielded are not accidentally junked and disposed of.
type Wieldable struct {
	Attribute
	slots []string
}

// Some interfaces we want to make sure we implement
var (
	_ has.Wieldable = &Wieldable{}
	_ has.Vetoes    = &Wieldable{}
	_ has.Slotable  = &Wieldable{}
)

// NewWieldable returns a new Wieldable attribute initialised with the passed
// Body slot references. Any Thing with a Wieldable attribute can be wielded by
// a player or mobile provided they have the specified Body slots available.
func NewWieldable(slots ...string) *Wieldable {
	return &Wieldable{Attribute{}, slots}
}

// FindWieldable searches the attributes of the specified Thing for attributes
// that implement has.Wieldable returning the first match it finds or a
// *Wieldable typed nil otherwise.
func FindWieldable(t has.Thing) has.Wieldable {
	return t.FindAttr((*Wieldable)(nil)).(has.Wieldable)
}

// Is returns true if passed attribute implements a wieldable item else false.
func (*Wieldable) Is(a has.Attribute) bool {
	_, ok := a.(has.Wieldable)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (w *Wieldable) Found() bool {
	return w != nil
}

// Unmarshal is used to turn the passed data into a new Wieldable attribute.
func (*Wieldable) Unmarshal(data []byte) has.Attribute {

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
	return NewWieldable(slots...)
}

// Marshal returns a tag and []byte that represents the receiver.
func (w *Wieldable) Marshal() (tag string, data []byte) {

	iSlots := make(map[string]int, len(w.slots))    // Slots + integer quantities
	sSlots := make(map[string]string, len(w.slots)) // Slots + string quantities

	for _, slot := range w.slots {
		iSlots[slot]++
	}

	for slot, count := range iSlots {
		sSlots[slot] = strconv.Itoa(count)
	}

	return "wieldable", encode.PairList(sSlots, '→')
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (w *Wieldable) Dump(node *tree.Node) *tree.Node {
	slots := []byte{}
	if len(w.slots) > 0 {
		for _, slot := range w.slots {
			slots = strconv.AppendQuote(append(slots, ", "...), slot)
		}
		slots = slots[2:]
	}
	return node.Append("%p %[1]T - slots: %d [%s]", w, len(w.slots), slots)
}

// IsWieldable returns true.
func (w *Wieldable) IsWieldable() bool {
	return true
}

// Slots returns the Body slot references that need to be available to wield
// the Thing. The returned slice should not be modified.
func (w *Wieldable) Slots() []string {
	return w.slots
}

// Check will veto the JUNK command if the Thing is currently wielded.
func (w *Wieldable) Check(actor has.Thing, cmd ...string) has.Veto {

	for _, cmd := range cmd {

		// If command not JUNK we won't veto
		if cmd != "JUNK" {
			continue
		}

		// If actor has no body it can't be wielding so we won't veto
		b := FindBody(actor)
		if !b.Found() {
			continue
		}

		if usage := b.Usage(w.Parent()); usage != "" {
			name := FindName(w.Parent()).TheName("something")
			return NewVeto(cmd, "You connot junk "+name+" while "+usage+" it.")
		}
	}
	return nil
}

// Copy returns a copy of the Wieldable receiver.
func (w *Wieldable) Copy() has.Attribute {
	if w == nil {
		return (*Wieldable)(nil)
	}
	return NewWieldable(w.Slots()...)
}
