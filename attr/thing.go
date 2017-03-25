// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/text"

	"fmt"
	"log"
)

// Thing is a container for Attributes. Everything in WolfMUD is constructed by
// creating a Thing and then adding Attributes to it which implement specific
// functionality.
type Thing struct {
	attrs []has.Attribute
}

// Some interfaces we want to make sure we implement
var (
	_ has.Thing = &Thing{}
)

// NewThing returns a new Thing initialised with the specified Attributes.
// Attributes can also be dynamically modified using Add and Remove methods.
func NewThing(a ...has.Attribute) *Thing {
	t := &Thing{}
	t.Add(a...)
	return t
}

// Free is used to clean-up/release references to all Attribute for a Thing.
// When a Thing is finished with calling Free helps the garbage collector to
// reclaim objects. It can also help to break cyclic references that could
// prevent garbage collection.
func (t *Thing) Free() {
	if t == nil {
		return
	}
	for _, a := range t.attrs {
		a.Free()
	}
	t.Remove(t.attrs...)
	t.attrs = nil
}

// Add is used to add the passed Attributes to a Thing. When an Attribute is
// added its parent is set to reference the Thing it was added to. This allows
// an Attribute to find and query the parent Thing about other Attributes the
// Thing may have.
func (t *Thing) Add(a ...has.Attribute) {
	for _, a := range a {
		a.SetParent(t)
		t.attrs = append(t.attrs, a)
	}
}

// Remove is used to remove the passed Attributes from a Thing. When an
// Attribute is removed its parent it set to nil. There is no indication if an
// Attribute cannot actually be removed.
func (t *Thing) Remove(a ...has.Attribute) {
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(t.attrs) - 1; j >= 0; j-- {
			if a[i] == t.attrs[j] {
				copy(t.attrs[j:], t.attrs[j+1:])
				t.attrs[len(t.attrs)-1] = nil
				t.attrs = t.attrs[:len(t.attrs)-1]
				break
			}
		}
	}
}

// Attrs returns all of the Attributes a Thing has as a slice of has.Attribute.
// This is commonly used to range over all of the Attributes of a Thing instead
// of using a finder for a specific type of Attribute.
func (t *Thing) Attrs() []has.Attribute {
	return t.attrs
}

// ignoredFields is a list of known field names that should be ignored by
// Unmarshal as there is no corresponding Attribute to unmarshal the field's
// data. If the field isn't ignored we just get extra warnings in the log when
// unmarshaling is attempted.
var ignoredFields = text.Dictionary("ref", "location", "zonelinks")

// Unmarshal unmarshals a Thing from a recordjar record containing all of the
// Attribute to be added. The recno is the record number in the recordjar for
// this record. It is passed so that we can give informative messages if errors
// are found. If the record number is not known -1 should be passed instead.
func (t *Thing) Unmarshal(recno int, record recordjar.Record) {

	var (
		m  has.Marshaler
		a  has.Attribute
		ok bool
	)

	// Go through the fields in the record
	for field, data := range record {

		// Some known fields without attributes or marshalers we don't want to
		// try and unmarshal so we ignore.
		if ignoredFields.Contains(field) {
			continue
		}

		// Look for a marshaler for the field name
		if m, ok = internal.Marshalers[field]; !ok {
			if recno == -1 {
				log.Printf("Unknown attribute: %s", field)
			} else {
				log.Printf("[Record: %d] Unknown attribute: %s", recno, field)
			}
			continue
		}

		// Unmarshal the data into an Attribute and add it to the Thing as long as
		// the returned, unmarshaled Attribute is not an untyped nil.
		if a = m.Unmarshal(data); a != nil {
			t.Add(a)
		}
	}

	return
}

func (t *Thing) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T %d attributes:", t, len(t.attrs)))
	for _, a := range t.attrs {
		for _, a := range a.Dump() {
			buff = append(buff, DumpFmt("%s", a))
		}
	}
	return buff
}

func DumpFmt(format string, args ...interface{}) string {
	return "  " + fmt.Sprintf(format, args...)
}

// Copy returns a copy of the Thing receiver. The copy will be made recursively
// copying all associated Attribute and Thing.
func (t *Thing) Copy() has.Thing {
	if t == nil {
		return (*Thing)(nil)
	}
	na := make([]has.Attribute, len(t.attrs), len(t.attrs))
	for i, a := range t.attrs {
		na[i] = a.Copy()
	}
	return NewThing(na...)
}

// SetOrigins updates the origin for the Thing to its containing Inventory and
// recursivly sets the origins for the content of a Thing's Inventory if it has
// one.
func (t *Thing) SetOrigins() {
	if t == nil {
		return
	}

	// Set our origin to that of the parent Inventory
	if l := FindLocate(t); l.Found() {
		if i := FindInventory(l.Where().Parent()); i.Found() {
			l.SetOrigin(i)
		}
	}

	// Find our Inventory
	i := FindInventory(t)
	if !i.Found() {
		return
	}

	// Set the origin for items in our Inventory
	for _, t := range append(i.Contents(), i.Narratives()...) {
		if l := FindLocate(t); l.Found() {
			l.SetOrigin(i)
		}
		t.SetOrigins()
	}
}
