// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Register marshaler for Locate attribute.
func init() {
	internal.AddMarshaler((*Locate)(nil), "locate")
}

// Locate implements an attribute that refers to the Inventory of where
// something is. When a Thing is added to an Inventory a Locate attribute will
// be added automatically if the Thing does not already have one. When a Thing
// is added to or removed from an Inventory the Locate.SetWhere method is
// called to update the reference. See inventory.Add for more details.
// Locate also records the initial starting position or origin of a Thing.
type Locate struct {
	Attribute
	where  has.Inventory
	origin has.Inventory
}

// Some interfaces we want to make sure we implement
var (
	_ has.Locate = &Locate{}
)

// NewLocate returns a new Locate attribute initialised to refer to the passed
// Inventory. Passing nil is a valid reference and is usually treated as being
// nowhere.
func NewLocate(i has.Inventory) *Locate {
	l := &Locate{Attribute{}, nil, nil}
	l.SetWhere(i)
	return l
}

// FindLocate searches the attributes of the specified Thing for attributes
// that implement has.Locate returning the first match it finds or a *Locate
// typed nil otherwise.
func FindLocate(t has.Thing) has.Locate {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Locate); ok {
			return a
		}
	}
	return (*Locate)(nil)
}

// Found returns false if the receiver is nil otherwise true.
func (l *Locate) Found() bool {
	return l != nil
}

// Unmarshal is used to turn the passed data into a new Locate attribute. At
// the moment Locate attributes are created internally so return an untyped nil
// so we get ignored.
func (*Locate) Unmarshal(data []byte) has.Attribute {
	return nil
}

func (l *Locate) Dump() []string {
	origin := FindName(l.origin.Parent()).Name("Nowhere")
	where := FindName(l.where.Parent()).Name("Nowhere")
	return []string{DumpFmt("%p %[1]T -> Origin: %p %s, Where: %p %s", l, l.origin, origin, l.where, where)}
}

// Where returns the Inventory the parent Thing is in. Returning nil is a
// valid reference and is usually treated as being nowhere. The current
// Inventory is set by calling SetWhere.
func (l *Locate) Where() (where has.Inventory) {
	if l != nil {
		where = l.where
	}
	return
}

// Origin return the initial starting Inventory that a Thing is placed into.
func (l *Locate) Origin() (origin has.Inventory) {
	if l != nil {
		origin = l.origin
	}
	return
}

// SetWhere is used to set the Inventory containing the parent Thing. Passing
// nil is a valid reference and is usually treated as being nowhere. The
// current reference can be retrieved by calling Where.
//
// NOTE: This is called automatically by the Inventory Add and Remove methods.
func (l *Locate) SetWhere(i has.Inventory) {
	if l != nil {
		l.where = i
	}
}

// SetOrigin is use to specify the initial starting Inventory that a Thing is
// placed into.
func (l *Locate) SetOrigin(i has.Inventory) {
	if l != nil {
		l.origin = i
	}
}

// Copy returns a copy of the Locate receiver.
func (l *Locate) Copy() has.Attribute {
	if l == nil {
		return (*Locate)(nil)
	}
	return NewLocate(l.where)
}
