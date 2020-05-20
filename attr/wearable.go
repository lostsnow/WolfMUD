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
)

// Register marshaler for wearable attribute.
func init() {
	internal.AddMarshaler((*Wearable)(nil), "wearable")
}

// Wearable implements an attribute for specifying body slots required when
// wearing a Thing. Wearable will veto the JUNK command so that items being
// worn are not accidentally junked and disposed of.
type Wearable struct {
	Attribute
	slots []string
}

// Some interfaces we want to make sure we implement
var (
	_ has.Wearable = &Wearable{}
	_ has.Vetoes   = &Wearable{}
	_ has.Slotable = &Wearable{}
)

// NewWearable returns a new Wearable attribute initialised with the passed
// Body slot references. Any Thing with a Wearable attribute can be worn by
// a player or mobile provided they have the specified Body slots available.
func NewWearable(slots ...string) *Wearable {
	return &Wearable{Attribute{}, slots}
}

// FindWearable searches the attributes of the specified Thing for attributes
// that implement has.Wearable returning the first match it finds or a
// *Wearable typed nil otherwise.
func FindWearable(t has.Thing) has.Wearable {
	return t.FindAttr((*Wearable)(nil)).(has.Wearable)
}

// Is returns true if passed attribute implements a wearable item else false.
func (*Wearable) Is(a has.Attribute) bool {
	_, ok := a.(has.Wearable)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (w *Wearable) Found() bool {
	return w != nil
}

// Unmarshal is used to turn the passed data into a new Wearable attribute.
func (*Wearable) Unmarshal(data []byte) has.Attribute {

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
	return NewWearable(slots...)
}

// Marshal returns a tag and []byte that represents the receiver.
func (w *Wearable) Marshal() (tag string, data []byte) {

	iSlots := make(map[string]int, len(w.slots))    // Slots + integer quantities
	sSlots := make(map[string]string, len(w.slots)) // Slots + string quantities

	for _, slot := range w.slots {
		iSlots[slot]++
	}

	for slot, count := range iSlots {
		sSlots[slot] = strconv.Itoa(count)
	}

	return "wearable", encode.PairList(sSlots, '→')
}

func (w *Wearable) Dump() (buff []string) {
	return []string{DumpFmt("%p %[1]T: slots: %v", w, w.slots)}
}

// IsWearable return true
func (w *Wearable) IsWearable() bool {
	return true
}

// Slots returns the Body slot references that need to be available to wear
// the Thing. The returned slice should not be modified.
func (w *Wearable) Slots() []string {
	return w.slots
}

// Check will veto the JUNK command if the Thing is currently worn.
func (w *Wearable) Check(actor has.Thing, cmd ...string) has.Veto {

	for _, cmd := range cmd {

		// If command not JUNK we won't veto
		if cmd != "JUNK" {
			continue
		}

		// If actor has no body it can't be wearing so we won't veto
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

// Copy returns a copy of the Wearable receiver.
func (w *Wearable) Copy() has.Attribute {
	if w == nil {
		return (*Wearable)(nil)
	}
	return NewWearable(w.Slots()...)
}
