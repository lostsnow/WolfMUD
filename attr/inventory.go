// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
)

const (
	reclaimFactor = 2  // is capacity > length * reclaimFactor
	reclaimBuffer = 4  // only reclaim if gain more than reclaimBuffer
	crowdSize     = 10 // If inventory has more player than this it's a crowd
)

// Inventory implements an attribute for container inventories. The most common
// container usage is for locations and rooms as well as actual containers like
// bags, boxes and inventories for mobiles. WolfMUD does not actually define a
// specific type for locations. Locations are simply Things that have Inventory
// and usually Exits attributes.
//
// BUG(diddymus): Inventory capacity is not implemented yet.
type Inventory struct {
	Attribute
	contents    []has.Thing
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
	c := make([]has.Thing, len(t))

	// Shallow copy only - interface headers or pointers
	copy(c, t)

	return &Inventory{Attribute{}, c, 0, internal.NewBRL()}
}

// FindInventory searches the attributes of the specified Thing for attributes
// that implement has.Inventory returning the first match it finds or nil
// otherwise.
func FindInventory(t has.Thing) has.Inventory {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Inventory); ok {
			return a
		}
	}
	return nil
}

func (i *Inventory) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T Lock ID: %d, %d items:", i, i.LockID(), len(i.contents)))
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
	i.contents = append(i.contents, t)
	if a := FindLocate(t); a != nil {
		a.SetWhere(i)
	}

	// TODO: Need to check for players or mobiles
	if a := FindPlayer(t); a != nil {
		i.playerCount++
	}
}

// Remove tries to take the specified Thing from the Inventory. If the Thing is
// removed successfully it is returned otherwise nil is returned. If the Thing
// needs to know where it is - because it implements the has.Locate interface -
// we update where the Thing is to nil as it is now nowhere.
//
// TODO: The reclaim factor and buffer should be tunable via the configuration.
//
// TODO: A slice is fine for conveniance and simplicity but maybe a linked list
// would be better?
func (i *Inventory) Remove(t has.Thing) has.Thing {
	for j, c := range i.contents {
		if c == t {
			if a := FindLocate(t); a != nil {
				a.SetWhere(nil)
			}
			i.contents[j] = nil
			i.contents = append(i.contents[:j], i.contents[j+1:]...)

			// If we are using less than length*reclaimFactor of the slice's capacity
			// and the difference is more than reclaimBuffer 'shrink' the slice by
			// allocating a new slice of the exact size needed. The reclaimBuffer
			// stops us shrinking small buffers all the time where the gain is
			// minimal.
			if cap(i.contents)-(len(i.contents)*reclaimFactor) > reclaimBuffer {
				i.contents = append([]has.Thing(nil), i.contents[:]...)
			}

			// TODO: Need to check for players or mobiles
			if a := FindPlayer(t); a != nil {
				i.playerCount--
			}

			return c
		}
	}
	return nil
}

// Search returns the first Inventory Thing that matches the alias passed. If
// no matches are found nil is returned.
func (i *Inventory) Search(alias string) has.Thing {
	for _, c := range i.contents {
		if a := FindAlias(c); a != nil {
			if a.HasAlias(alias) {
				return c
			}
		}
	}
	return nil
}

// Contents returns a 'copy' of the Inventory contents. That is a copy of the
// slice containing has.Thing interface headers. Therefore the Inventory
// contents may be indirectly manipulated through the copy but changes to the
// actual slice are not possible - use the Add and Remove methods instead.
func (i *Inventory) Contents() []has.Thing {
	l := make([]has.Thing, len(i.contents))
	copy(l, i.contents)
	return l
}

// List returns a string describing the contents of an Inventory. The format of
// the string is dependant on the number of items. If the Inventory is empty:
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
func (i *Inventory) List() string {
	buff := make([]byte, 0, 1024)

	switch len(i.contents) {
	case 0:
		return "It is empty."
	case 1:
		buff = append(buff, "It contains "...)
	default:
		buff = append(buff, "It contains:\n  "...)
	}

	mark := len(buff)

	for _, c := range i.contents {
		if a := FindName(c); a != nil {
			if len(buff) > mark {
				buff = append(buff, "\n  "...)
			}
			buff = append(buff, a.Name()...)
		}
	}

	// End single item sentence with a fullstop.
	if len(i.contents) == 1 {
		buff = append(buff, "."...)
	}

	return string(buff)
}

func (i *Inventory) Crowded() bool {
	return i.playerCount > crowdSize
}
