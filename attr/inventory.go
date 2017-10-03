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
// A Thing in an Inventory may be disabled and taken out of play or enabled and
// put back into play. A disabled Thing is inaccessible to players but is still
// covered by the Inventory lock. This is so that any Thing can always be
// covered by a lock in an Inventory. An example usage of disabling/enabling a
// Thing is when an item is cleaned up and needs to be reset. In this case the
// clean up event triggering would cause the Thing to be moved to its origin
// Inventory and then disabling the Thing would cause it to go out of play.
// When the reset event triggers the Thing would be enabled and brought back
// into play.
//
// TODO: A slice for contents is fine for convenience and simplicity but maybe
// a linked list would be better? This would possibly save reslicing in Remove.
//
// BUG(diddymus): Inventory capacity is not implemented yet.
type Inventory struct {
	Attribute
	contents    []has.Thing
	split       int
	disabled    []has.Thing
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
	i := &Inventory{
		Attribute: Attribute{},
		contents:  make([]has.Thing, 0, len(t)),
		disabled:  []has.Thing{},
		BRL:       internal.NewBRL(),
	}

	for _, t := range t {
		i.AddDisabled(t)
		i.Enable(t)
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
	buff = append(buff, DumpFmt("%p %[1]T Lock ID: %d, %d items (players: %d, split: %d, disabled: %d):", i, i.LockID(), len(i.contents)+len(i.disabled), i.playerCount, i.split, len(i.disabled)))
	for _, i := range i.contents {
		for _, i := range i.Dump() {
			buff = append(buff, DumpFmt("%s", i))
		}
	}
	for _, i := range i.disabled {
		for _, i := range i.Dump() {
			buff = append(buff, DumpFmt("%s", i))
		}
	}
	return buff
}

// Move removes a Thing from the receiver Inventory and puts it into the
// 'where' Inventory. After the move the Thing's Locate attribute will be
// updated to reflect the new Inventory it is in.
func (i *Inventory) Move(t has.Thing, where has.Inventory) {

	if t == nil {
		return
	}

	// If where to move to is not an actual *Inventory we can't manipulate the
	// contents directly and so have to take the (very) slow path...
	to, ok := where.(*Inventory)
	if !ok {
		i.Disable(t)
		i.RemoveDisabled(t)
		to.AddDisabled(t)
		to.Enable(t)
		return
	}

	n := FindNarrative(t).Found()
	p := FindPlayer(t).Found()
	found := false

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

			// If Thing removed was a Narrative adjust Narrative/Thing split
			if n {
				i.split--
			}

			found = true
		}
	}

	if !found {
		return
	}

	// If Thing added was a Narrative move it to the front of the slice
	// and adjust the Narrative/Thing split.
	if n {
		to.contents = append(to.contents, nil)
		copy(to.contents[1:], to.contents[0:])
		to.contents[0] = t
		to.split++
	}

	// If Thing added not a Narrative just append it to the end of the slice
	if !n {
		to.contents = append(to.contents, t)
	}

	// TODO: Need to check for players or mobiles
	if p {
		to.playerCount++
	}

	// Update Where attribute on Thing with 'to' Inventory
	FindLocate(t).SetWhere(to)

	return
}

// AddDisabled adds a Thing to an Inventory marking at as being initially out
// of play. The Locate attribute of the Thing will be updated to reference the
// Inventory the Thing is put into. If the Thing does not have a Locate
// attribute one will be added.
//
//  TODO: AddDisable is only required because if we use Inventory.Add followed
//  by an Inventory.Disable the Add would trigger events and loop. This needs
//  to be cleaned up, possibly by making Add/Remove act on the disabled slice
//  only. This would mean a Thing can only be added to or removed from the
//  world when disabled and once in the world and enabled can only be moved
//  from Inventory to Inventory.
func (i *Inventory) AddDisabled(t has.Thing) {
	i.disabled = append(i.disabled, t)

	// If Locate attribute found update it, otherwise add a new one
	if l := FindLocate(t); l.Found() {
		l.SetWhere(i)
		return
	}
	t.Add(NewLocate(i))
}

// RemoveDisabled takes a disabled Thing from an Inventory.
//
// NOTE: Once the Thing is removed it will no longer be under a lock. Ideally
// once a Thing is removed Thing.Free should be called to release the Thing for
// garbage collection.
func (i *Inventory) RemoveDisabled(t has.Thing) {
	for j, a := range i.disabled {
		if a == t {
			copy(i.disabled[j:], i.disabled[j+1:])
			i.disabled[len(i.disabled)-1] = nil
			i.disabled = i.disabled[:len(i.disabled)-1]
			return
		}
	}
}

// Enabled marks a Thing in an Inventory as being in play.
func (i *Inventory) Enable(t has.Thing) {
	for j, a := range i.disabled {
		if a == t {
			copy(i.disabled[j:], i.disabled[j+1:])
			i.disabled[len(i.disabled)-1] = nil
			i.disabled = i.disabled[:len(i.disabled)-1]

			// If Thing added was a Narrative move it to the front of the slice and
			// adjust the Narrative/Thing split. If Thing added not a Narrative just
			// append it to the end of the slice
			if FindNarrative(t).Found() {
				i.contents = append(i.contents, nil)
				copy(i.contents[1:], i.contents[0:])
				i.contents[0] = t
				i.split++
			} else {
				i.contents = append(i.contents, t)
			}

			if FindPlayer(t).Found() {
				i.playerCount++
			}

			return
		}
	}
}

// Disable marks a Thing in an Inventory as being out of play.
func (i *Inventory) Disable(t has.Thing) {
	for j, c := range i.contents {
		if c == t {
			copy(i.contents[j:], i.contents[j+1:])
			i.contents[len(i.contents)-1] = nil
			i.contents = i.contents[:len(i.contents)-1]
			i.disabled = append(i.disabled, t)
			FindLocate(t).SetWhere(i)

			// If Thing removed was a Narrative adjust Narrative/Thing split
			if FindNarrative(t).Found() {
				i.split--
			}

			if FindPlayer(t).Found() {
				i.playerCount--
			}

			return
		}
	}
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

func (i *Inventory) Disabled() []has.Thing {
	if i == nil {
		return []has.Thing{}
	}

	l := make([]has.Thing, len(i.disabled))
	copy(l, i.disabled[:])
	return l
}

// List returns a string describing the non-narrative contents of an Inventory.
// The layout of the description returned is dependant on the number of items.
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

// Crowded tests to see if an Inventory has so many players in it that it is
// considered crowded. If the Inventory is considered crowded true is returned
// otherwise false. An Inventory is considered crowded if there are more than
// config.Inventory.CrowdSize players in it.
func (i *Inventory) Crowded() (crowded bool) {
	if i != nil {
		crowded = i.playerCount > config.Inventory.CrowdSize
	}
	return
}

// Players returns true if there are any players in the Inventory else false.
func (i *Inventory) Players() bool {
	return i.playerCount > 0
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
		c := a.Copy()
		ni.AddDisabled(c)
		ni.Enable(c)
	}
	for _, a := range i.disabled {
		ni.AddDisabled(a.Copy())
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
	for x, t := range i.disabled {
		i.disabled[x] = nil
		t.Free()
	}
	i.Attribute.Free()
}

// Carried returns true if putting an item into the Inventory would result in
// it being carried by a player, otherwise false. The Inventory can be the
// player's actual Inventory or the Inventory of a container (checked
// recursively) in the player's inventory.
//
// TODO: Need to check for players or mobiles
func (i *Inventory) Carried() bool {
	if i == nil {
		return false
	}

	var where has.Inventory = i

	for where != nil {
		p := where.Parent()
		if FindPlayer(p).Found() {
			return true
		}
		where = FindLocate(p).Where()
	}

	return false
}

// Outermost returns the top level inventory in an Inventory hierarchy.
func (i *Inventory) Outermost() has.Inventory {

	var (
		p has.Thing
		l has.Locate
		w has.Inventory
	)

	if p = i.Parent(); p == nil {
		return i
	}
	if l = FindLocate(p); !l.Found() {
		return i
	}
	if w = l.Where(); w == nil || !w.Found() {
		return i
	}

	return w.Outermost()
}
