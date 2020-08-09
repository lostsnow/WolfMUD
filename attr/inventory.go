// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"log"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Inventory attribute.
func init() {
	internal.AddMarshaler((*Inventory)(nil), "inventory", "inv")
}

// Inventory implements an attribute for container inventories. The most common
// container usage is for locations and rooms as well as actual containers like
// bags, boxes and inventories for players and mobiles. WolfMUD does not
// actually define a specific type for locations. Locations are simply Things
// that have an Exits attribute.
//
// Any Thing added to an Inventory will automatically be assigned a Locate
// attribute. A locate attribute is simply a back reference to the Inventory a
// Thing is in. This enables a Thing to work out where it is.
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
// BUG(diddymus): Inventory capacity is not implemented yet.
type Inventory struct {
	Attribute
	players    *list
	contents   *list
	narratives *list
	disabled   *list
	internal.BRL
}

// list implements a simple double linked list optimised for Inventories
// holding has.Thing items. The head and tail are both sentinal nodes. A list
// can be walked using:
//
//    for n := list.head.next; n.item != nil; n = n.next {
//      // forwards
//    }
//
//    for n := list.tail.prev; n.item != nil; n = n.prev {
//      // reverse
//    }
//
// New items are always added at the head of the list.
type list struct {
	head *node
	tail *node
	len  int
}

// newList sets up a new, empty list with sentinal head and tail nodes. The
// head sentinal node should always have prev = nil, the tail sentinal next =
// nil.
func newList() *list {
	l := &list{head: &node{}, tail: &node{}}
	l.head.next = l.tail
	l.tail.prev = l.head
	return l
}

// free releases the head and tail from a list. The list should be empty before
// being freed. Calling free helps the garbage collector as otherwise the head
// is linked to the tail and the tail linked to the head, in an empty list,
// creating a cyclic reference between the head and tail sentinal nodes.
func (l *list) free() {
	// unlink head and tail
	l.head.next = nil
	l.tail.prev = nil
	// clear head and tail
	l.head = nil
	l.tail = nil
}

// node represents an item in a list.
type node struct {
	prev *node
	next *node
	item has.Thing
}

// add appends an item to the head of the receiver list.
func (l *list) add(t has.Thing) {
	l.head.next.prev = &node{l.head, l.head.next, t}
	l.head.next = l.head.next.prev
	l.len++
}

// remove walks a list and removes the specified thing if found.
func (l *list) remove(t has.Thing) bool {
	for n := l.head.next; n.next != nil; n = n.next {
		if n.item != t {
			continue
		}
		n.next.prev, n.prev.next = n.prev, n.next
		n.prev, n.next, n.item = nil, nil, nil
		l.len--
		return true
	}
	return false
}

// move unlinks the node containing the passed Thing from the receiver list and
// links it into the destination to list. This is a remove+add that reuses the
// list node avoiding throwing away the old node and creating a new one. If the
// remove fails false will be returned otherwise true.
func (l *list) move(t has.Thing, to *list) bool {
	for n := l.head.next; n.next != nil; n = n.next {
		if n.item != t {
			continue
		}
		n.next.prev, n.prev.next = n.prev, n.next
		l.len--

		n.prev, n.next = to.head, to.head.next
		to.head.next.prev = n
		to.head.next = n
		to.len++

		return true
	}
	return false
}

// list dumps a list for debugging
func (l *list) list(label string) {
	log.Printf("list for %s", label)
	for n := l.head; n.next != nil; n = n.next {
		log.Printf("list %12p, %12p, %12p, %s", n, n.prev, n.next, n.item)
	}
	log.Printf("list %12p, %12p, %12p, %s", l.tail, l.tail.prev, l.tail.next, l.tail.item)
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
var (
	_ has.Inventory = &Inventory{}
)

// NewInventory returns a new Inventory attribute initialised with the
// specified Things as initial contents. All of the Thing added will be enabled
// and in play - although the Thing itself may not be enabled and in play.
func NewInventory(t ...has.Thing) *Inventory {
	i := &Inventory{
		Attribute:  Attribute{},
		players:    newList(),
		contents:   newList(),
		narratives: newList(),
		disabled:   newList(),
		BRL:        internal.NewBRL(),
	}

	for _, t := range t {
		i.Add(t)
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
	return t.FindAttr((*Inventory)(nil)).(has.Inventory)
}

// Is returns true if passed attribute implements an inventory else false.
func (*Inventory) Is(a has.Attribute) bool {
	_, ok := a.(has.Inventory)
	return ok
}

// Unmarshal is used to turn the passed data into a new Inventory attribute.
func (*Inventory) Unmarshal(data []byte) has.Attribute {
	return NewInventory()
}

// Marshal returns a tag and []byte that represents the receiver.
func (i *Inventory) Marshal() (tag string, data []byte) {
	var refs []string
	for _, list := range []*list{i.contents, i.narratives} {
		for n := list.head.next; n.next != nil; n = n.next {
			refs = append(refs, n.item.UID())
		}
	}
	for n := i.disabled.head.next; n.next != nil; n = n.next {
		refs = append(refs, "!"+n.item.UID())
	}
	return "inventory", encode.KeywordList(refs)
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (i *Inventory) Dump(node *tree.Node) *tree.Node {
	node = node.Append("%p %[1]T - lock ID: %d, items: %d",
		i,
		i.LockID(),
		i.players.len+i.contents.len+i.narratives.len+i.disabled.len,
	)

	lists := node.Branch()

	for _, list := range []struct {
		label string
		*list
	}{
		{"players", i.players},
		{"contents", i.contents},
		{"narratives", i.narratives},
		{"disabled", i.disabled},
	} {
		lists.Append("%p %[1]T - (%s) len: %d", list.list, list.label, list.len)
		branch := lists.Branch()
		for n := list.head.next; n.next != nil; n = n.next {
			n.item.Dump(branch)
		}
	}

	return node
}

// Move removes an enabled Thing from the receiver Inventory and puts it into
// the 'where' Inventory. After the move the Thing's Locate attribute will be
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
		i.Remove(t)
		to.Add(t)
		to.Enable(t)
		return
	}

	switch {
	case i.players.move(t, to.players):
	case i.contents.move(t, to.contents):
	case i.narratives.move(t, to.narratives):
	default:
		return
	}

	// Update Where attribute on Thing with 'to' Inventory
	FindLocate(t).SetWhere(to)

	return
}

// Add puts a Thing into an Inventory marking at as being initially out
// of play. The Locate attribute of the Thing will be updated to reference the
// Inventory the Thing is put into. If the Thing does not have a Locate
// attribute one will be added. The Thing may be enabled and put in play by
// calling Enable.
func (i *Inventory) Add(t has.Thing) {
	i.disabled.add(t)
	if l := FindLocate(t); l.Found() {
		l.SetWhere(i)
		return
	}
	t.Add(NewLocate(i))
}

// Remove takes a disabled Thing out of an Inventory.
//
// NOTE: Once the Thing is removed it will no longer be under a lock. Ideally
// once a Thing is removed Thing.Free should be called to release the Thing for
// garbage collection.
func (i *Inventory) Remove(t has.Thing) {
	FindLocate(t).SetWhere(nil)
	i.disabled.remove(t)
}

// Enabled marks a Thing in an Inventory as being in play.
func (i *Inventory) Enable(t has.Thing) {
	switch {
	case FindPlayer(t).Found():
		i.disabled.move(t, i.players)
	case FindNarrative(t).Found():
		i.disabled.move(t, i.narratives)
	default:
		i.disabled.move(t, i.contents)
	}
}

// Disable marks a Thing in an Inventory as being out of play.
func (i *Inventory) Disable(t has.Thing) {
	switch {
	case i.players.len != 0 && FindPlayer(t).Found():
		i.players.move(t, i.disabled)
	case i.narratives.len != 0 && FindNarrative(t).Found():
		i.narratives.move(t, i.disabled)
	default:
		i.contents.move(t, i.disabled)
	}
}

// Search returns the first Inventory Thing that matches the alias passed. If
// no matches are found nil is returned.
func (i *Inventory) Search(alias string) has.Thing {
	if i == nil {
		return nil
	}

	for _, list := range []*list{i.players, i.contents, i.narratives} {
		for n := list.tail.prev; n.prev != nil; n = n.prev {
			if FindAlias(n.item).HasAlias(alias) {
				return n.item
			}
		}
	}
	return nil
}

// Players returns a list of Players in an Inventory. The players may be
// indirectly manipulated through the slice. Players should be added to, or
// removed from the Inventory using the Add and Remove methods.
//
// See also the Contents, Narratives and Everything methods.
func (i *Inventory) Players() (l []has.Thing) {
	if i == nil {
		return
	}
	l = make([]has.Thing, 0, i.players.len)
	for n := i.players.tail.prev; n.prev != nil; n = n.prev {
		l = append(l, n.item)
	}
	return
}

// Contents returns a list of items in an Inventory. The items may be
// indirectly manipulated through the slice. Items should be added to, or
// removed from the Inventory using the Add and Remove methods.
//
// See also the Players, Narratives and Everything methods.
func (i *Inventory) Contents() (l []has.Thing) {
	if i == nil {
		return
	}
	l = make([]has.Thing, 0, i.contents.len)
	for n := i.contents.tail.prev; n.prev != nil; n = n.prev {
		l = append(l, n.item)
	}
	return
}

// Narratives returns a list of narrative items in an Inventory. The items may
// be indirectly manipulated through the slice. Items should be added to, or
// removed from the Inventory using the Add and Remove methods.
//
// See also the Players, Contents and Everything methods.
func (i *Inventory) Narratives() (l []has.Thing) {
	if i == nil {
		return
	}
	l = make([]has.Thing, 0, i.narratives.len)
	for n := i.narratives.tail.prev; n.prev != nil; n = n.prev {
		l = append(l, n.item)
	}
	return
}

// Everything returns a list of all Players, Content and Narratives in an
// Inventory. The items may be indirectly manipulated through the slice. Items
// should be added to, or removed from the Inventory using the Add and Remove
// methods.
//
// See also the Players, Contents and Narratives methods.
func (i *Inventory) Everything() (l []has.Thing) {
	if i == nil {
		return
	}
	l = make([]has.Thing, 0, i.players.len+i.contents.len+i.narratives.len)
	for _, list := range []*list{i.players, i.contents, i.narratives} {
		for n := list.tail.prev; n.prev != nil; n = n.prev {
			l = append(l, n.item)
		}
	}
	return
}

func (i *Inventory) Disabled() (l []has.Thing) {
	if i == nil {
		return
	}
	l = make([]has.Thing, 0, i.disabled.len)
	for n := i.disabled.tail.prev; n.prev != nil; n = n.prev {
		l = append(l, n.item)
	}
	return
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

	switch i.players.len + i.contents.len {
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

	for _, list := range []*list{i.players, i.contents} {
		for n := list.tail.prev; n.prev != nil; n = n.prev {
			if len(buff) > mark {
				buff = append(buff, "\n  "...)
			}
			buff = append(buff, FindName(n.item).Name("Something")...)
		}
	}

	// End single item sentence with a fullstop.
	if i.players.len+i.contents.len == 1 {
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
		crowded = i.players.len > config.Inventory.CrowdSize
	}
	return
}

// Occupied returns true if there is at least one player in the Inventory.
func (i *Inventory) Occupied() bool {
	return i.players.len > 0
}

// Empty returns true if there are no non-Narrative items else false.
func (i *Inventory) Empty() bool {
	if i != nil {
		return i.players.len+i.contents.len == 0
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
	for _, list := range []*list{i.contents, i.narratives} {
		for n := list.head.next; n.next != nil; n = n.next {
			c := n.item.DeepCopy()
			ni.Add(c)
			ni.Enable(c)
		}
	}
	for n := i.disabled.head.next; n.next != nil; n = n.next {
		ni.Add(n.item.DeepCopy())
	}
	return ni
}

// Free recursively calls Free on all of it's content when the Inventory
// attribute is freed.
func (i *Inventory) Free() {
	if i == nil {
		return
	}

	for _, list := range []*list{i.contents, i.narratives, i.disabled} {
		for list.head.next.next != nil {
			list.head.next.item.Free()
			list.remove(list.head.next.item)
		}
		list.free()
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
