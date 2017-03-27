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

// Register marshaler for Inventory attribute.
func init() {
	internal.AddMarshaler((*Inventory)(nil), "inventory", "inv")
}

// Inventory implements an attribute for container inventories. The most common
// container usage is for locations and rooms as well as actual containers like
// bags, boxes and inventories for mobiles. WolfMUD does not actually define a
// specific type for locations. Locations are simply Things that have an Exits
// attribute.
//
// Any Thing added to an Inventory will automatically be assigned a Locate
// attribute. A locate attribute is simply a back reference to the Inventory a
// Thing is in. This enables a Thing to work out where it is.
//
// NOTE: The contents slice is split into two parts. Things with a Narrative
// attribute are added to the beginning of the slice. All other Things are
// appended to the end of the slice. Which items are narrative and which are
// not is tracked by split:
//
//	narratives := contents[:split]
//	other := contents[split:]
//
//	countNarratives := split
//	countOther := len(contents) - split
//
// For a complete description of narratives see the Narrative attribute type.
//
// TODO: A slice for contents is fine for convenience and simplicity but maybe
// a linked list would be better? This would possibly save reslicing in Remove.
//
// BUG(diddymus): Inventory capacity is not implemented yet.
type Inventory struct {
	Attribute
	contents    []has.Thing
	split       int
	playerCount int
	internal.BRL
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
var (
	_ has.Inventory = &Inventory{}
)

// NewInventory returns a new Inventory attribute initialised with the
// specified Things as initial contents.
func NewInventory(t ...has.Thing) *Inventory {
	c := make([]has.Thing, 0, len(t))
	i := &Inventory{Attribute{}, c, 0, 0, internal.NewBRL()}

	for _, t := range t {
		i.Add(t)
	}

	return i
}

// Found returns false if the receiver is nil otherwise true.
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

// Unmarshal is used to turn the passed data into a new Inventory attribute.
func (*Inventory) Unmarshal(data []byte) has.Attribute {
	return NewInventory()
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

// Add puts a Thing into an Inventory. If the Thing does not have a Locate
// attribute one will be added automatically, otherwise the existing Locate
// attribute will be updated. On success Add will return the Thing actually
// added to the inventory - which may not be the Thing passed in, it may be a
// copy. It is therefore important to use the Thing returned after calling Add.
// On failure Add returns nil.
func (i *Inventory) Add(t has.Thing) has.Thing {
	if i == nil {
		return nil
	}
	return (*Inventory)(nil).Move(t, i)
}

// Remove takes a Thing from an Inventory. On success Remove will return the
// Thing actually removed from the inventory - which may not be the Thing
// passed in, it may be a copy. It is therefore important to use the Thing
// returned after calling Remove. If Remove fails it will return nil.
func (i *Inventory) Remove(t has.Thing) has.Thing {
	if i == nil {
		return nil
	}
	return i.Move(t, nil)
}

// Move removes a Thing from one Inventory and puts it into another Inventory.
// On success Move will return the Thing moved - which may not be the Thing
// passed in, it may be a copy. It is therefore important to use the Thing
// returned after calling Move. If Move fails it will return nil.
//
// If the receiver is a *Inventory typed nil the Thing will only be added to an
// inventory. If the to Inventory is nil the Thing will only be removed from
// the reveiver Inventory. In both cases the Thing's Locate attribute will be
// updated or one added if missing.
func (i *Inventory) Move(t has.Thing, to has.Inventory) has.Thing {

	if t == nil {
		return t
	}

	n := FindNarrative(t).Found()
	p := FindPlayer(t).Found()
	l := FindLocate(t)
	found := false

	if i == nil {
		goto ADD
	}

	for j, c := range i.contents {
		if c == t {
			copy(i.contents[j:], i.contents[j+1:])
			i.contents[len(i.contents)-1] = nil
			i.contents = i.contents[:len(i.contents)-1]

			// If we are using less than half of the slice's capacity and the
			// difference is more than config.Inventory.Compact 'shrink' the slice by
			// allocating a new slice of the exact size needed. The value of
			// config.Inventory.Compact stops us shrinking small buffers all the time
			// where the gain is minimal.
			if l, c := len(i.contents), cap(i.contents); (c - l - l) >= config.Inventory.Compact {
				i.contents = append(make([]has.Thing, 0, l), i.contents...)
			}

			// TODO: Need to check for players or mobiles
			if p {
				i.playerCount--
			}

			// If Thing removed was a Narrative adjust split
			if n {
				i.split--
			}

			// If not a player check if removing a Thing triggers a re-spawning.
			// Players don't respawn but they do move from location to location a lot
			// which would cause needless calls to Spawn.
			if !p {
				if s := FindReset(t).Spawn(); s != nil {
					t = s
				}
			}

			found = true
		}
	}

	if !found {
		return nil
	}

ADD:

	To, ok := to.(*Inventory)

	if to == nil {
		goto UPDATE
	}

	// If to is not an actual *Inventory have to take the slow path
	if !ok {
		return to.Add(t)
	}

	// If Thing added was a narrative move it to the front of the slice otherwise
	// just append it onto the end. Adjust split if Thing is narrative.
	if n {
		To.contents = append(To.contents, nil)
		copy(To.contents[1:], To.contents[0:])
		To.contents[0] = t
		To.split++
	} else {
		To.contents = append(To.contents, t)
	}

	// TODO: Need to check for players or mobiles
	if p {
		To.playerCount++
	}

UPDATE:

	// Give thing a locate attribute if it doesn't have one, else just update it
	if !l.Found() {
		t.Add(NewLocate(To))
	} else {
		l.SetWhere(To)
	}

	return t
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

// Narratives returns a 'copy' of the Inventory narrative contents. That is a
// copy of the slice containing has.Thing interface headers. Therefore the
// Inventory narratives may be indirectly manipulated through the copy but
// changes to the actual slice are not possible - use the Add and Remove
// methods instead.
func (i *Inventory) Narratives() []has.Thing {
	if i == nil {
		return []has.Thing{}
	}
	l := make([]has.Thing, i.split)
	copy(l, i.contents[:i.split])
	return l
}

// List returns a string describing the non-narrative contents of an Inventory.
// The layout of the dscription returned is dependant on the number of items.
// If the Inventory is empty and the Parent Thing has a narrative attribute we
// return nothing. Otherwise if the Inventory is empty we return:
//
//	It is empty.
//
// A single item only we return:
//
//	It contains xxx.
//
// For multiple items we return:
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
		if FindNarrative(i.Parent()).Found() {
			return ""
		}
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
		crowded = i.playerCount > config.Inventory.CrowdSize
	}
	return
}

// Empty returns true if there are no non-Narrative items else false.
func (i *Inventory) Empty() bool {
	if i != nil {
		return len(i.contents)-i.split == 0
	}
	return true
}

// Copy returns a copy of the Inventory receiver. The copy will be made
// recursively copying the complete content of the Inventory as well.
//
// NOTE: There are no checks made for cyclic references which could send us
// into infinite recursion. However cyclic references should be prevented by
// the zone loader. See zones.isParent function.
func (i *Inventory) Copy() has.Attribute {
	if i == nil {
		return (*Inventory)(nil)
	}
	ni := NewInventory()
	for _, a := range i.contents {
		ni.Add(a.Copy())
	}
	return ni
}

// Free recursively calls Free on all of it's content when the Inventory
// attribute is freed.
func (i *Inventory) Free() {
	if i == nil {
		return
	}
	for x, t := range i.contents {
		i.contents[x] = nil
		t.Free()
	}
	i.Attribute.Free()
}
