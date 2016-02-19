// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Inventory implements an attribute for container inventories. The most common
// container usage is for locations and rooms as well as actual containers like
// bags, boxes and inventories for mobiles. WolfMUD does not actually define a
// specific type for locations. Locations are simply Things that have Inventory
// and usually Exits attributes.
//
// For a complete description of narratives see the Narrative attribute type.
//
// NOTE: The contents slice is split into two parts. Things with a Narrative
// attribute are added to the begenning of the slice. All other Things are
// appended to the end of the slice. Which items are narrative and which are
// not is tracked by split:
//
//	narattives := contents[:split]
//	other := contents[split:]
//
//	countNarratives := split
//	countOther := len(contents) - split
//
// BUG(diddymus): Inventory capacity is not implemented yet.
type Inventory struct {
	Attribute
	contents    []has.Thing
	split       int
	playerCount uint64
	internal.BRL
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
var (
	_ has.Inventory = &Inventory{}
)

// NewInventory returns a new Inventory attribute initialised with the
// specified Things as initial contents.
//
// BUG(diddymus): NewInventory should use proper copies of the Things passed.
// Until Attribute and Thing implement a Copy method we can't do that.
// Implementing a Copy method instead of building a reflect deep copy is more
// desirable as it will allow us to fine tune exactly what is copied and how it
// is copied when duplicating a Thing.
func NewInventory(t ...has.Thing) *Inventory {
	c := make([]has.Thing, 0, len(t))
	i := &Inventory{Attribute{}, c, 0, 0, internal.NewBRL()}

	for _, t := range t {
		i.Add(t)
	}

	return i
}

// Found returns false if the receiver is nil otherwise true. This is a utility
// method that can be chained with FindInventory to easily check if an
// Inventory attribute was found.
func (i *Inventory) Found() bool {
	return i != nil
}

// FindInventory searches the attributes of the specified Thing for attributes
// that implement has.Inventory returning the first match it finds or a
// *Inventory typed nil otherwise.
func FindInventory(t has.Thing) has.Inventory {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Inventory); ok {
			return a
		}
	}
	return (*Inventory)(nil)
}

func (i *Inventory) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T Lock ID: %d, %d items (split: %d):", i, i.LockID(), len(i.contents), i.split))
	for _, i := range i.contents {
		for _, i := range i.Dump() {
			buff = append(buff, DumpFmt("%s", i))
		}
	}
	return buff
}

// Add puts the specified Thing into the Inventory. If the Thing needs to know
// where it is - because it implements the has.Locate interface - we update
// where the Thing is to point to the Inventory.
func (i *Inventory) Add(t has.Thing) {
	if i == nil {
		return
	}

	// If Thing added was a narrative move it to the front of the slice otherwise
	// just append it onto the end. Adjust split if Thing is narrative.
	if FindNarrative(t).Found() {
		i.contents = append(i.contents, nil)
		copy(i.contents[1:], i.contents[0:])
		i.contents[0] = t
		i.split++
	} else {
		i.contents = append(i.contents, t)
		FindLocate(t).SetWhere(i)
	}

	// TODO: Need to check for players or mobiles
	if FindPlayer(t).Found() {
		i.playerCount++
	}
}

// Remove tries to take the specified Thing from the Inventory. If the Thing is
// removed successfully it is returned otherwise nil is returned. If the Thing
// needs to know where it is - because it implements the has.Locate interface -
// we update where the Thing is to nil as it is now nowhere.
//
// TODO: A slice is fine for conveniance and simplicity but maybe a linked list
// would be better?
func (i *Inventory) Remove(t has.Thing) has.Thing {
	if i == nil {
		return nil
	}

	for j, c := range i.contents {
		if c == t {
			FindLocate(t).SetWhere(nil)
			i.contents[j] = nil
			i.contents = append(i.contents[:j], i.contents[j+1:]...)

			// If we are using less than length*reclaimFactor of the slice's capacity
			// and the difference is more than reclaimBuffer 'shrink' the slice by
			// allocating a new slice of the exact size needed. The reclaimBuffer
			// stops us shrinking small buffers all the time where the gain is
			// minimal.
			if cap(i.contents)-(len(i.contents)*config.ReclaimFactor) > config.ReclaimBuffer {
				i.contents = append([]has.Thing(nil), i.contents[:]...)
			}

			// TODO: Need to check for players or mobiles
			if FindPlayer(t).Found() {
				i.playerCount--
			}

			// If Thing removed was a Narrative adjust split
			if FindNarrative(t).Found() {
				i.split--
			}

			return c
		}
	}
	return nil
}

// Search returns the first Inventory Thing that matches the alias passed. If
// no matches are found nil is returned.
func (i *Inventory) Search(alias string) has.Thing {
	if i == nil {
		return nil
	}

	for _, c := range i.contents {
		if FindAlias(c).HasAlias(alias) {
			return c
		}
	}
	return nil
}

// Contents returns a 'copy' of the Inventory non-narrative contents. That is a
// copy of the slice containing has.Thing interface headers. Therefore the
// Inventory contents may be indirectly manipulated through the copy but
// changes to the actual slice are not possible - use the Add and Remove
// methods instead.
func (i *Inventory) Contents() []has.Thing {
	if i == nil {
		return []has.Thing{}
	}
	l := make([]has.Thing, len(i.contents)-i.split)
	copy(l, i.contents[i.split:])
	return l
}

// List returns a string describing the non-narrative contents of an Inventory.
// The format of the string is dependant on the number of items. If the
// Inventory is empty:
//
//	It is empty.
//
// A single item only:
//
//	It contains xxx.
//
// Multiple items:
//
//	It contains:
//		Item
//		Item
//		Item
//		...
//
// If the inventory cannot be listed an empty string will be returned.
func (i *Inventory) List() string {
	if i == nil {
		return ""
	}

	buff := make([]byte, 0, 1024)

	switch len(i.contents) - i.split {
	case 0:
		return "It is empty."
	case 1:
		buff = append(buff, "It contains "...)
	default:
		buff = append(buff, "It contains:\n  "...)
	}

	mark := len(buff)

	for _, c := range i.contents[i.split:] {
		if len(buff) > mark {
			buff = append(buff, "\n  "...)
		}
		buff = append(buff, FindName(c).Name("Something")...)
	}

	// End single item sentence with a fullstop.
	if len(i.contents)-i.split == 1 {
		buff = append(buff, "."...)
	}

	return string(buff)
}

func (i *Inventory) Crowded() (crowded bool) {
	if i != nil {
		crowded = i.playerCount > config.CrowdSize
	}
	return
}

// Count returns the total number of objects, number of Narratives and number
// of items in the specified Inventory.
func (i *Inventory) Count() (L, N, I int) {
	if i != nil {
		return len(i.contents), i.split, len(i.contents) - i.split
	}
	return 0, 0, 0
}
