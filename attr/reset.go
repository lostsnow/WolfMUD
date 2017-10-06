// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"time"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/event"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"
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
	return &Reset{Attribute{}, after, jitter, spawn, nil}
}

// FindReset searches the attributes of the specified Thing for attributes
// that implement has.Reset returning the first match it finds or a *Reset
// typed nil otherwise.
func FindReset(t has.Thing) has.Reset {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Reset); ok {
			return a
		}
	}
	return (*Reset)(nil)
}

// Found returns false if the receiver is nil otherwise true.
func (r *Reset) Found() bool {
	return r != nil
}

// Unmarshal is used to turn the passed data into a new Reset attribute.
func (*Reset) Unmarshal(data []byte) has.Attribute {
	r := NewReset(0, 0, false)
	for _, pairs := range recordjar.Decode.PairList(data) {
		field, data := pairs[0], []byte(pairs[1])
		switch field {
		case "AFTER":
			r.after = recordjar.Decode.Duration(data)
		case "JITTER":
			r.jitter = recordjar.Decode.Duration(data)
		case "SPAWN":
			r.spawn = recordjar.Decode.Boolean(data)
		}
	}
	return r
}

func (r *Reset) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T After: %s Jitter: %s Spawn: %t", r, r.after, r.jitter, r.spawn))
	buff = append(buff, DumpFmt("  %p %[1]T", r.Cancel))
	return
}

// Copy returns a copy of the Reset receiver. The copy will not inherit any
// pending Reset events.
func (r *Reset) Copy() has.Attribute {
	if r == nil {
		return (*Reset)(nil)
	}
	return NewReset(r.after, r.jitter, r.spawn)
}

// Reset schedules a reset of the parent Thing. If there is already a reset
// event pending it will be cancelled and a new one queued.
func (r *Reset) Reset() {
	if r == nil {
		return
	}

	// Register for reset cancelling any outstanding resets
	if r.Cancel != nil {
		close(r.Cancel)
	}
	r.Cancel = event.Queue(r.Parent(), "$RESET", r.after, r.jitter)
}

// Spawn returns a non-spawnable copy of a Thing and schedules the original
// Thing to reset if Reset.spawn is true. Otherwise it returns the original
// Thing.
func (r *Reset) Spawn() has.Thing {

	// If not spawnable just exit
	if r == nil {
		return nil
	}

	// If not spawnable return original thing
	if !r.spawn {
		return r.Parent()
	}

	// Make a copy of original Thing, update origins of the copy to point to any
	// copied Inventories
	p := r.Parent()
	c := p.Copy()
	c.SetOrigins()

	// Disable original Thing and register a reset for it
	o := FindLocate(p).Origin()
	o.Disable(p)
	r.Reset()

	// Remove reset attribute from copied Thing
	R := FindReset(c)
	c.Remove(R)
	R.Free()

	// Set origin of copy to nil so it will be disposed of when cleaned up as it
	// is the original that respawns. Then add copy back into the world.
	l := FindLocate(c)
	l.SetOrigin(nil)
	l.Where().Add(c)
	l.Where().Enable(c)

	return c
}

// Free makes sure references are nil'ed and channels closed when the Reset
// attribute is freed.
func (r *Reset) Free() {
	if r == nil {
		return
	}
	if r.Cancel != nil {
		close(r.Cancel)
		r.Cancel = nil
	}
	r.Attribute.Free()
}
