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
// something is. When a Thing changes the Inventory it is contained in and has
// a Locate attribute SetWhere should be called to update the reference. This
// attribute only needs to be added to things that need to know where they are.
// For example a player needs to know where they are so that they can move
// themselves.
type Locate struct {
	Attribute
	where has.Inventory
}

// Some interfaces we want to make sure we implement
var (
	_ has.Locate = &Locate{}
)

// NewLocate returns a new Locate attribute initialised to refer to the passed
// Inventory. Passing nil is a valid reference and is usually treated as being
// nowhere.
func NewLocate(i has.Inventory) *Locate {
	l := &Locate{Attribute{}, nil}
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
	name := ""
	if w := l.Where(); w != nil {
		name = FindName(w.Parent()).Name("Somewhere")
	}
	return []string{DumpFmt("%p %[1]T -> %p %s", l, l.where, name)}
}

// Where returns the Inventory where 'we' are. Returning nil is a valid
// reference and is usually treated as being nowhere. The current reference
// should be set by calling SetWhere.
func (l *Locate) Where() (where has.Inventory) {
	if l != nil {
		where = l.where
	}
	return
}

// SetWhere is used to set the Inventory where 'we' are. Passing nil is a valid
// reference and is usually treated as being nowhere. The current reference can
// be retrieved by calling Where.
//
// NOTE: This is called automatically by the Inventory Add and Remove methods.
func (l *Locate) SetWhere(i has.Inventory) {
	if l != nil {
		l.where = i
	}
}