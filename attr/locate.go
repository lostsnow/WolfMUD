// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Locate implements an attribute that refers to the Thing containing the
// inventory of where something is. When a Thing changes the Inventory it is
// contained in and has a Locate attribute SetWhere should be called to update
// the reference. This attribute only needs to be added to things that need to
// know where they are. For example a player needs to know where they are so
// that they can move themselves.
//
// TODO: Need to check implications of changing the where field type from
// has.Thing to has.Inventory - this could simplify things and save us from a
// number of FindInventory lookups for many commands.
type Locate struct {
	Attribute
	where has.Thing
}

// Some interfaces we want to make sure we implement
var (
	_ has.Locate = &Locate{}
)

// NewLocate returns a new Locate attribute initialised to refer to the passed
// thing. Passing nil is a valid reference and is usually treated as being
// nowhere.
func NewLocate(t has.Thing) *Locate {
	l := &Locate{Attribute{}, nil}
	l.SetWhere(t)
	return l
}

// FindLocate searches the attributes of the specified Thing for attributes
// that implement has.Locate returning the first match it finds or nil
// otherwise.
func FindLocate(t has.Thing) has.Locate {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Locate); ok {
			return a
		}
	}
	return nil
}

func (l *Locate) Dump() []string {
	name := ""
	if w := l.Where(); w != nil {
		if a := FindName(w); a != nil {
			name = a.Name()
		}
	}
	return []string{DumpFmt("%p %[1]T -> %p %s", l, l.where, name)}
}

// Where returns the Thing where 'we' are. To be precise it returns the Thing
// with the Inventory attribute that contains the Thing that has this Locate
// attribute. Returning nil is a valid reference and is usually treated as
// being nowhere. The current reference can be set by calling SetWhere.
func (l *Locate) Where() has.Thing {
	return l.where
}

// SetWhere is used to set the Thing where 'we' are. It should be the Thing
// with the Inventory attribute that contains the Thing that has this Locate
// attribute. Passing nil is a valid reference and is usually treated as being
// nowhere. The current reference can be retrieved by calling Where.
//
// NOTE: This is called automatically by the Inventory Add and Remove methods.
//
// TODO: Should we be checking that w has an inventory if we are being placed
// there? Switching to Inventories instead of Things would solve this - see
// todo for Locate type about switching to Inventories.
func (l *Locate) SetWhere(w has.Thing) {
	l.where = w
}
