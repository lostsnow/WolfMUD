// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/has"

	"fmt"
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
	for _, a := range a {
		for k, v := range t.attrs {
			if v == a {
				t.attrs[k] = nil
				a.SetParent(nil)
				t.attrs = append(t.attrs[:k], t.attrs[k+1:]...)
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
