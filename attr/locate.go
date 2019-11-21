// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"sync"

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
// Concurrent access of a Locate attribute is safe.
type Locate struct {
	Attribute

	rwmutex sync.RWMutex
	where   has.Inventory
	origin  has.Inventory
}

// Some interfaces we want to make sure we implement
var (
	_ has.Locate = &Locate{}
)

// NewLocate returns a new Locate attribute initialised to refer to the passed
// Inventory. Passing nil is a valid reference and is usually treated as being
// nowhere.
func NewLocate(i has.Inventory) *Locate {
	l := &Locate{Attribute: Attribute{}}
	l.SetWhere(i)
	return l
}

// FindLocate searches the attributes of the specified Thing for attributes
// that implement has.Locate returning the first match it finds or a *Locate
// typed nil otherwise.
func FindLocate(t has.Thing) has.Locate {
	return t.FindAttr((*Locate)(nil)).(has.Locate)
}

// Is returns true if passed attribute implements locate else false.
func (*Locate) Is(a has.Attribute) bool {
	_, ok := a.(has.Locate)
	return ok
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

// Marshal returns a tag and []byte that represents the receiver. In this case
// we return empty values as the Locate attribute is not persisted.
func (*Locate) Marshal() (string, []byte) {
	return "", []byte{}
}

func (l *Locate) Dump() (buf []string) {
	origin := "Nowhere"
	where := "Nowhere"
	l.rwmutex.RLock()
	if l.origin != nil && l.origin.Found() {
		origin = FindName(l.origin.Parent()).Name("no name!")
	}
	if l.where != nil && l.where.Found() {
		where = FindName(l.where.Parent()).Name("no name!")
	}
	buf = append(buf, DumpFmt("%p %[1]T -> Origin: %p %s, Where: %p %s", l, l.origin, origin, l.where, where))
	l.rwmutex.RUnlock()
	return
}

// Where returns the Inventory the parent Thing is in. Returning nil is a
// valid reference and is usually treated as being nowhere. The current
// Inventory is set by calling SetWhere.
func (l *Locate) Where() (where has.Inventory) {
	if l != nil {
		l.rwmutex.RLock()
		where = l.where
		l.rwmutex.RUnlock()
	}
	return
}

// Origin return the initial starting Inventory that a Thing is placed into.
func (l *Locate) Origin() (origin has.Inventory) {
	if l != nil {
		l.rwmutex.RLock()
		origin = l.origin
		l.rwmutex.RUnlock()
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
		l.rwmutex.Lock()
		l.where = i
		l.rwmutex.Unlock()
	}
}

// SetOrigin is use to specify the initial starting Inventory that a Thing is
// placed into.
func (l *Locate) SetOrigin(i has.Inventory) {
	if l != nil {
		l.rwmutex.Lock()
		l.origin = i
		l.rwmutex.Unlock()
	}
}

// Copy returns a copy of the Locate receiver.
func (l *Locate) Copy() has.Attribute {
	if l == nil {
		return (*Locate)(nil)
	}
	l.rwmutex.RLock()
	nl := NewLocate(l.where)
	l.rwmutex.RUnlock()
	return nl
}

// Free makes sure references are nil'ed when the Locate attribute is freed.
func (l *Locate) Free() {
	if l == nil {
		return
	}
	l.rwmutex.Lock()
	l.where = nil
	l.origin = nil
	l.rwmutex.Unlock()
	l.Attribute.Free()
}
