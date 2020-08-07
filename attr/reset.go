// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"log"
	"time"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/event"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Reset attribute.
func init() {
	internal.AddMarshaler((*Reset)(nil), "reset")
}

// Reset implements an Attribute for resetting or respawning Things in the game
// world. When a Thing is disposed of in the game world it may need to be reset
// and placed back into the game world at it's initial starting position after
// a delay period has elapsed. Otherwise the world quickly becomes empty with
// little for players to do. Some Thing's may respawn instead of resetting.
// When a Thing respawns another copy of the Thing is placed into the game
// world after a period of time when the Thing is taken. For both cases the
// delay period will be between Reset.after and Reset.after+Reset.jitter. If a
// Thing is being reset/respawned and is in its delay period the Reset.Cancel
// channel will be non-nil and the reset/respawn may be aborted by closing the
// channel. If Reset.spawn is true the Thing is respawnable otherwise it is
// resettable. Items that should just be removed when disposed of should not
// have a Reset attribute.
type Reset struct {
	Attribute
	after  time.Duration
	jitter time.Duration
	spawn  bool
	dueAt  time.Time     // Time a queued event is expected to fire
	dueIn  time.Duration // Time remaining for a suspended event
	event.Cancel
}

// Some interfaces we want to make sure we implement
var (
	_ has.Reset = &Reset{}
)

// Reset implements an attribute for resetting or respawning Things and putting
// them back into the game world. The after and jitter Duration set the delay
// period to between after and after+jitter for when a Thing is reset or
// respawned. If spawn is true the Thing will respawn otherwise it will reset.
func NewReset(after time.Duration, jitter time.Duration, spawn bool) *Reset {
	return &Reset{Attribute{}, after, jitter, spawn, time.Time{}, 0, nil}
}

// FindReset searches the attributes of the specified Thing for attributes
// that implement has.Reset returning the first match it finds or a *Reset
// typed nil otherwise.
func FindReset(t has.Thing) has.Reset {
	return t.FindAttr((*Reset)(nil)).(has.Reset)
}

// Is returns true if passed attribute implements a reset else false.
func (*Reset) Is(a has.Attribute) bool {
	_, ok := a.(has.Reset)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (r *Reset) Found() bool {
	return r != nil
}

// Unmarshal is used to turn the passed data into a new Reset attribute.
func (*Reset) Unmarshal(data []byte) has.Attribute {
	r := NewReset(0, 0, false)
	for field, data := range decode.PairList(data) {
		data := []byte(data)
		switch field {
		case "AFTER":
			r.after = decode.Duration(data)
		case "JITTER":
			r.jitter = decode.Duration(data)
		case "SPAWN":
			r.spawn = decode.Boolean(data)
		case "DUE-IN", "DUE_IN":
			r.dueIn = decode.Duration(data)
		default:
			log.Printf("Reset.unmarshal unknown attribute: %q: %q", field, data)
		}
	}
	return r
}

// Marshal returns a tag and []byte that represents the receiver.
func (r *Reset) Marshal() (tag string, data []byte) {
	tag = "reset"
	pairs := map[string]string{
		"after":  string(encode.Duration(r.after)),
		"jitter": string(encode.Duration(r.jitter)),
		"spawn":  string(encode.Boolean(r.spawn)),
	}

	switch {
	case r.Cancel != nil:
		pairs["due_In"] = string(encode.Duration(time.Until(r.dueAt)))
	case r.dueIn > 0:
		pairs["due_In"] = string(encode.Duration(r.dueIn))
	}

	data = encode.PairList(pairs, '→')
	return
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (r *Reset) Dump(node *tree.Node) *tree.Node {
	node = node.Append(
		"%p %[1]T - after: %s, jitter: %s, spawn: %t",
		r, r.after, r.jitter, r.spawn,
	)

	var due, source string
	if r.Cancel != nil {
		due = time.Until(r.dueAt).String()
		source = "at"
	} else {
		due = r.dueIn.String()
		source = "in"
	}
	node.Branch().Append("%p %[1]T - due: %s, source: %s", r.Cancel, due, source)

	return node
}

// Copy returns a copy of the Reset receiver. If the Reset event is currently
// queued it will be suspended in the returned copy.
func (r *Reset) Copy() has.Attribute {
	if r == nil {
		return (*Reset)(nil)
	}
	nr := NewReset(r.after, r.jitter, r.spawn)
	if r.Cancel != nil {
		nr.dueIn = time.Until(r.dueAt)
	} else {
		nr.dueIn = r.dueIn
	}
	return nr
}

// schedule queues a Reset event to occur after the given delay has passed. The
// delay will be between 'after' and 'after+jitter'. If the Reset event is
// already queued it will be cancelled and a new one queued.
func (r *Reset) schedule(after, jitter time.Duration) {
	r.Abort()

	// Schedule event, for a $RESET the actor is where the reset will take place.
	what := r.Parent()
	actor := FindLocate(what).Origin().Parent()

	r.dueIn = 0
	r.Cancel, r.dueAt = event.Queue(actor, "$RESET "+what.UID(), after, jitter)
}

// Reset schedules a Reset event. If the Reset event is already queued it will
// be cancelled and a new one queued.
func (r *Reset) Reset() {
	if r != nil {
		r.schedule(r.after, r.jitter)
	}
}

// Suspend a queued Reset event, or do nothing if event not queued.
func (r *Reset) Suspend() {
	if r == nil {
		return
	}

	if r.Cancel != nil {
		close(r.Cancel)
		r.Cancel = nil
		r.dueIn = time.Until(r.dueAt)
		r.dueAt = time.Time{}
	}
}

// Resume a suspended Reset event, or do nothing if event not suspended.
func (r *Reset) Resume() {
	if r != nil && r.dueIn != 0 {
		r.schedule(r.dueIn, 0)
	}
}

// Abort a queued Reset event, or do nothing if event not queued.
func (r *Reset) Abort() {
	if r == nil {
		return
	}

	if r.Cancel != nil {
		close(r.Cancel)
		r.Cancel = nil
		r.dueAt = time.Time{}
	}
}

// Pending returns true if there is a Reset event pending, else false. Use
// with caution as this could introduce a race between checking the state and
// acting on it as the event could fire between the two actions.
func (r *Reset) Pending() bool {
	return r.Cancel != nil
}

// Spawn returns a non-spawnable copy of a spawnable Thing and schedules the
// original Thing to reset if Reset.spawn is true. Otherwise it returns nil.
//
// If a new item is spawned then the Inventory of the original is processed.
// Unique and non-spawnable items are moved from the original to the copy.
// Copies of spawnable content are made and the original spawnable content is
// disabled and a reset scheduled. This processing is recursive.
func (r *Reset) Spawn() has.Thing {

	if r == nil || !r.spawn {
		return nil
	}

	// Spawnable so make a copy of original Thing, disable original and register
	// original for a reset
	p := r.Parent()
	c := p.Copy()
	FindLocate(p).Origin().Disable(p)
	r.Reset()

	// Remove reset attribute from copy and clear origin - only originals respawn
	R := FindReset(c)
	c.Remove(R)
	R.Free()
	l := FindLocate(c)
	l.SetOrigin(nil)

	r.spawnInventory(p, c)

	// Add copy back into the world
	l.Where().Add(c)
	l.Where().Enable(c)

	return c
}

// spawnInventory recursively spawns the content of the Inventory from one
// Thing to another. Spawning will either move non-spawnable items or copy
// spawnable items. If a disabled item is copied it's reset is rescheduled.
//
// Note that we can't use Thing.DeepCopy as we need to selectively move or copy
// items to the spawned Thing, and possibly reschedule a Reset for copied
// disabled things.
func (r *Reset) spawnInventory(from, to has.Thing) {

	// If original has no Inventory nothing to do
	fromInv := FindInventory(from)
	if !fromInv.Found() {
		return
	}

	toInv := FindInventory(to)

	for _, t := range fromInv.Contents() {
		if !FindReset(t).Spawnable() {
			fromInv.Move(t, toInv)
			continue
		}
		c := t.Copy()
		FindLocate(c).SetOrigin(toInv)
		r.spawnInventory(t, c)
		toInv.Add(c)
		toInv.Enable(c)
	}

	for _, t := range fromInv.Disabled() {
		if !FindReset(t).Spawnable() {
			fromInv.Move(t, toInv)
			continue
		}
		c := t.Copy()
		FindLocate(c).SetOrigin(toInv)
		r.spawnInventory(t, c)
		FindReset(c).Resume()
		toInv.Add(c)
	}
}

// Spawnable returns true if the parent Thing is spawnable else false.
func (r *Reset) Spawnable() bool {
	return r != nil && r.spawn
}

// Unique returns true if item is considered unique else false. For an item to
// be unique it must be resetable and must not be spawnable.
//
// NOTE: An item without a reset is technically not unique as it is the
// byproduct of an item spawning and hence a copy of that item.
func (r *Reset) Unique() bool {
	return r != nil && !r.spawn
}

// Free makes sure references are nil'ed and queued events aborted when the
// Reset attribute is freed.
func (r *Reset) Free() {
	if r == nil {
		return
	}
	r.Abort()
	r.Attribute.Free()
}
