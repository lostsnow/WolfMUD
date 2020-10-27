// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"strconv"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Body attribute.
func init() {
	internal.AddMarshaler((*Body)(nil), "body")
}

// Body implements an attribute representing the body of a Thing which in turn
// represents the body slots available for holding, wearing and wielding items.
//
// NOTE: A Body attribute and its associated slots are not automatically kept
// in sync with a player's or mobile's Inventory. Should an item be removed or
// discarded the Body should also be updated, if needed, to reflect the change.
// At the moment this mainly effects the DROP and PUT commands. The QUIT
// command also needs to update the Body when non-collectable items are
// disposed of. The JUNK command is currently vetoed to stop accidental junking
// of in use items - if that changes the JUNK command will also need to update
// the Body.
//
// TODO(diddymus): Currently there is no relationship specified between body
// slots. For example, if a hand is missing then the fingers and thumb should
// also be missing but this is not handled automatically yet.
//
// BUG(diddymus): If there are multiple slots with the same name available, for
// example two HAND slots, only one will be reported. For example if you are
// holding a knife and dagger and try to wield a sword it will report one of:
//
//  - You cannot wield the sword while also using a knife.
//  - You cannot wield the sword while also using a dagger.
//
// Message depends on order of usage. Ideally this should report:
//
//  - You cannot wield the sword while also using a knife and a dagger.
//
// Maybe that should be 'a knife OR a dagger'? If you try to wield a two handed
// staff the message is correct:
//
//  - You cannot wield the staff while also using a knife and a dagger.
//
// I think the fix for this is to make UsedBy smarter.
type Body struct {
	Attribute
	slots []slot
}

// slot represents a specific body slot that can be used to hold, wear or wield
// an item. The granularity of slots is unrestricted. For example there could
// be a single TORSO slot or CHEST, BACK and SHOULDER slots. The only 'rule' is
// that the slots defining a body match those used when defining items. As an
// example a breastplate for a Body with a TORSO slot would be defined with
// 'Wearable: TORSO', while a Body with CHEST and BACK slots would define the
// breastplate as 'Wearable: CHEST BACK'. The latter would also allow for
// something like pauldons to be worn in the SHOULDER slots.
type slot struct {
	ref   string    // Slot reference e.g. "ARM", "LEG"
	used  has.Thing // Thing currently using slot or nil
	usage usageBits // See usageBits type and associated constants
}

// usageBits represents bit flags for information about a slot such as how a
// slot is being used.
type usageBits byte

// Constants for usageBits.
const (
	missing  usageBits = 1 << iota // Slot is missing
	wearing                        // Slot used to wear something
	wielding                       // Slot used to wield something
	holding                        // Slot used to hold something

	inUse = wearing | wielding | holding // Bit mask to check if slot used
)

// usageBitNames maps the usageBits constants to a meaningful name.
var usageBitNames = map[usageBits]string{
	missing:  "missing",
	wearing:  "wearing",
	wielding: "wielding",
	holding:  "holding",
}

// Some interfaces we want to make sure we implement
var (
	_ has.Body = &Body{}
)

// NewBody returns a Body attribute initialised with the slots specified by
// refs. The refs should contain the name of the slots to be made available for
// wearing, wielding or holding items. For example:
//
//  NewBody("HEAD", "TORSO", "ARM", "HAND", "ARM", "HAND", "LEG", "LEG")
//
// There are no restrictions on how detailed body composition is - if you want
// to define 10 fingers, 10 toes, eyebrows etc you can! The only 'rule' is that
// the slot references used for a body must match the references used by the
// Wearable and Wieldable attributes defined on items.
//
// If a body part is missing, for example a hand was cut off, it should be
// represented with a leading exclamation mark '!'. This then provides the
// possibility of regaining the body part through various means such as magic,
// prothetics or growing it back. For example:
//
//  NewBody("HEAD", "TORSO", "ARM", "!HAND", "ARM", "HAND", "LEG", "LEG")
//
// Without this record of a missing body part there is no way of knowing it was
// originally there in the first place.
func NewBody(refs ...string) *Body {
	b := &Body{Attribute{}, make([]slot, len(refs))}

	for x, ref := range refs {
		b.slots[x].ref = ref
		if ref[0] == '!' {
			b.slots[x].usage |= missing
			b.slots[x].ref = b.slots[x].ref[1:]
		}
	}

	return b
}

// FindBody searches the attributes of the specified Thing for attributes that
// implement has.Body returning the first match it finds or a *Body typed nil
// otherwise.
func FindBody(t has.Thing) has.Body {
	return t.FindAttr((*Body)(nil)).(has.Body)
}

// Is returns true if passed attribute implements a body else false.
func (*Body) Is(a has.Attribute) bool {
	_, ok := a.(has.Body)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (b *Body) Found() bool {
	return b != nil
}

// Unmarshal is used to turn the passed data into a new Body attribute.
func (*Body) Unmarshal(data []byte) has.Attribute {
	refs := []string{}
	var x int
	var err error

	for ref, count := range decode.PairList(data) {
		if x, err = strconv.Atoi(count); err != nil {
			x = 1
		}
		for ; x > 0; x-- {
			refs = append(refs, ref)
		}
	}
	return NewBody(refs...)
}

// Marshal returns a tag and []byte that represents the receiver.
func (b *Body) Marshal() (tag string, data []byte) {
	refs := make(map[string]int)
	for _, slot := range b.slots {
		ref := slot.ref
		if slot.usage&missing != 0 {
			ref = "!" + ref
		}
		if _, found := refs[ref]; !found {
			refs[ref] = 0
		}
		refs[ref]++
	}

	// Convert map[string]int to map[string]string for PairList
	slots := make(map[string]string, len(refs))
	for ref, x := range refs {
		slots[ref] = strconv.Itoa(x)
	}

	return "body", encode.PairList(slots, '→')
}

// save pre-marshal hook to make sure Holding, Wearing and Wielding attributes
// are present so that they are saved.
func (b *Body) save() {
	if b == nil {
		return
	}

	p := b.Parent()
	isHolding := FindHolding(p).Found()
	isWearing := FindWearing(p).Found()
	isWielding := FindWielding(p).Found()

	for _, slot := range b.slots {

		// Bail out early if we know we have all the attributes
		if isHolding && isWearing && isWielding {
			break
		}

		switch {
		case slot.usage&holding != 0 && !isHolding:
			isHolding = true
			p.Add(NewHolding())
		case slot.usage&wearing != 0 && !isWearing:
			isWearing = true
			p.Add(NewWearing())
		case slot.usage&wielding != 0 && !isWielding:
			isWielding = true
			p.Add(NewWielding())
		}
	}
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (b *Body) Dump(node *tree.Node) *tree.Node {
	node = node.Append("%p %[1]T", b)
	branch := node.Branch()
	for _, slot := range b.slots {
		slot.Dump(branch)
	}
	return node
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (s slot) Dump(node *tree.Node) *tree.Node {
	meanings := make([]string, 0, len(usageBitNames))
	for usage, name := range usageBitNames {
		if s.usage&usage != 0 {
			meanings = append(meanings, name)
		}
	}

	if s.used != nil {
		return node.Append(
			"%T - ref: %q, inuse: %p %q, usageBits: %08b %v",
			s, s.ref, s.used, FindName(s.used).Name("Something"), s.usage, meanings,
		)
	}

	return node.Append(
		"%T - ref: %q, usageBits: %08b %v",
		s, s.ref, s.usage, meanings,
	)
}

// Wield returns true if the Wieldable is successfully wielded else false. If
// successfully wielded Body slots will be allocated to the Wieldable and the
// slots marked as 'wielding', on failure no slots are allocated.
func (b *Body) Wield(w has.Wieldable) bool {
	return b.use(w, wielding)
}

// Wear returns true if the Wearable is successfully worn else false. If
// successfully worn Body slots will be allocated to the Wearable and the slots
// marked as 'wearing', on failure no slots are allocated.
func (b *Body) Wear(w has.Wearable) bool {
	return b.use(w, wearing)
}

// Hold returns true if the Holdable is successfully held else false. If
// successfully held Body slots will be allocated to the Holdable and the slots
// marked as 'holding', on failure no slots are allocated.
func (b *Body) Hold(h has.Holdable) bool {
	return b.use(h, holding)
}

// Remove the passed Thing from all Body slots allocated to it. Cleared Body
// slots will also have their usage cleared.
func (b *Body) Remove(t has.Thing) {
	if b == nil {
		return
	}
	for x, s := range b.slots {
		if s.used == t {
			b.slots[x].used = nil
			b.slots[x].usage &^= inUse
		}
	}
}

// RemoveAll frees up all available Body slots and stops using any Thing that
// are currently in use.
func (b *Body) RemoveAll() {
	if b == nil {
		return
	}
	for x := range b.slots {
		if b.slots[x].usage&inUse != 0 {
			b.slots[x].used = nil
			b.slots[x].usage &^= inUse
		}
	}
}

// Wielding returns a unique slice of Things, even if they take up more than
// one Body slot, currently being wielded by the Body. If nothing is being
// wielded an empty slice will be returned.
func (b *Body) Wielding() []has.Thing {
	return b.usedFor(wielding)
}

// Wearing returns a unique slice of Things, even if they take up more than one
// Body slot, currently being worn on the Body. If nothing is being worn an
// empty slice will be returned.
func (b *Body) Wearing() []has.Thing {
	return b.usedFor(wearing)
}

// Holding returns a unique slice of Things, even if they take up more than one
// Body slot, currently being held by the Body. If nothing is being held an
// empty slice will be returned.
func (b *Body) Holding() []has.Thing {
	return b.usedFor(holding)
}

// use allocates Body slots returned by the Slotable and sets the slot's usage.
// If all of the slots from Slotable can be allocated returns true else returns
// false and none of the slots are allocated.
func (b *Body) use(s has.Slotable, usage usageBits) bool {

	if b == nil {
		return false
	}

	what := s.Parent()
	refs := s.Slots()

nextRef:
	for _, ref := range refs {
		for x, s := range b.slots {
			if s.ref == ref && s.used == nil && s.usage&missing == 0 {
				b.slots[x].used = what
				b.slots[x].usage |= usage
				continue nextRef
			}
		}
		// Failed to assign all slots, so undo the slots we did assign.
		b.Remove(what)
		return false
	}
	return true
}

// Using returns true if at least one Body slot is allocated to the passed
// Thing else false.
func (b *Body) Using(t has.Thing) bool {
	if b == nil {
		return false
	}
	for _, s := range b.slots {
		if s.used == t {
			return true
		}
	}
	return false
}

// UsedBy returns a slice of unique Things, even if they are allocated to more
// than one slot, that are using the given slot references.
func (b *Body) UsedBy(refs []string) (usedBy []has.Thing) {
	if b == nil {
		return
	}

	work := make([]string, len(refs))
	copy(work, refs)
	unique := make(map[has.Thing]struct{})

	for _, s := range b.slots {
		if s.usage&missing != 0 {
			continue // Missing slots can't match
		}

		// See if slot matches any ref in the work list, if it does remove ref from
		// work list and break for next slot.
		for x, ref := range work {
			if s.ref == ref && s.used != nil {
				unique[s.used] = struct{}{}
				copy(work[x:], work[x+1:])
				work = work[:len(work)-1]
				break
			}
		}
	}
	for u := range unique {
		usedBy = append(usedBy, u)
	}
	return
}

// usedFor returns a unique slice of Things that are being used for a specific
// purpose, even if an item is using more than one Body slot. For example all
// items being worn. If there are no items that match the specified usage, or
// there is no Body attribute, an empty slice will be returned.
func (b *Body) usedFor(usage usageBits) []has.Thing {
	if b == nil {
		return []has.Thing{}
	}
	seen := make(map[has.Thing]struct{}, len(b.slots))
	for _, slot := range b.slots {
		if slot.usage&usage != 0 {
			if _, found := seen[slot.used]; !found {
				seen[slot.used] = struct{}{}
			}
		}
	}
	unique := make([]has.Thing, 0, len(seen))
	for what := range seen {
		unique = append(unique, what)
	}
	return unique
}

// Usage returns a string describing the Body slot usage for the passed Thing.
// For example 'wielding' if the Thing is being wielded. If the passed Thing is
// not currently allocated to any Body slots an empty string will be returned.
func (b *Body) Usage(t has.Thing) string {
	if b == nil {
		return ""
	}
	for _, s := range b.slots {
		if s.used == t {
			return usageBitNames[s.usage&inUse]
		}
	}
	return ""
}

// Has returns true if the Body has all of the slots specified by the passed
// refs else false. Has does not check if the slots are used or not, only if
// they are available.
func (b *Body) Has(refs []string) bool {
	if b == nil {
		return false
	}

	work := make([]string, len(refs))
	copy(work, refs)

	for _, s := range b.slots {
		if s.usage&missing != 0 {
			continue // Missing slots can't match
		}

		// See if slot matches any ref in the work list, if it does remove ref from
		// work list and break for next slot. If work list has one entry there is
		// no point removing it - we have found all the refs required so return
		// true.
		for x, ref := range work {
			if s.ref == ref {
				if len(work) == 1 {
					return true
				}
				copy(work[x:], work[x+1:])
				work = work[:len(work)-1]
				break
			}
		}
	}
	return false
}

// Copy returns a copy of the Body receiver including all of the slots and
// their current usage.
func (b *Body) Copy() has.Attribute {
	if b == nil {
		return (*Body)(nil)
	}
	body := NewBody()
	for _, s := range b.slots {
		slot := slot{s.ref, s.used, s.usage}
		body.slots = append(body.slots, slot)
	}
	return body
}

// Free makes sure Body slots are nil'ed when the Body attribute is freed.
func (b *Body) Free() {
	if b == nil {
		return
	}
	b.RemoveAll()
	b.slots = nil
	b.Attribute.Free()
}
