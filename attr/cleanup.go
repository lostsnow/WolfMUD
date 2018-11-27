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
)

// Register marshaler for Cleanup attribute.
func init() {
	internal.AddMarshaler((*Cleanup)(nil), "cleanup")
}

// Cleanup implements an Attribute for disposing of Things left laying around
// in the game world. When an item is dropped it will be cleaned up after a
// delay period has elapsed. Otherwise the world would get cluttered with
// items.
//
// The delay period is between Cleanup.after and Cleanup.after+Cleanup.jitter.
// If a Thing is being cleaned up and is in its delay period the Cleanup.Cancel
// channel will be non-nil and the clean up may be aborted by closing the
// channel or by calling Cleanup.Abort which will cancel clean up requests
// recursively for a Thing.
//
// SPECIFICS
//
// If an item is put into any Inventory and the item ends up not being carried
// by a player - either in their Inventory or in a container in their Inventory
// - and the receiving inventory has no parent Inventory that are already
// scheduled for clean up then the item is scheduled for a clean up. If the
// item has an Inventory (a container) the contents do not need to be scheduled
// for clean up recursively as everything will be cleaned up when the item
// itself is cleaned up. This is also why we don't schedule a clean up when
// putting an item in an Inventory where a parent Inventory is scheduled for a
// clean up already.
//
// If an item is removed from any Inventory any pending clean ups are
// cancelled. If the item has an Inventory its content - checked recursively -
// will have any pending clean ups cancelled. If we don't cancel pending clean
// ups recursively then putting an item into a container and then picking the
// container up would result in the item still being scheduled for a clean up,
// resulting in the item disappearing from the container.
type Cleanup struct {
	Attribute
	after  time.Duration
	jitter time.Duration
	event.Cancel
}

// Some interfaces we want to make sure we implement
var (
	_ has.Cleanup = &Cleanup{}
)

// NewCleanup returns a new Cleanup attribute initialised with the passed after
// and jitter durations. The after and jitter Duration set the delay period to
// between after and after+jitter for when a Thing is cleaned up after being
// dropped.
func NewCleanup(after time.Duration, jitter time.Duration) *Cleanup {
	return &Cleanup{Attribute{}, after, jitter, nil}
}

// FindCleanup searches the attributes of the specified Thing for attributes
// that implement has.Cleanup returning the first match it finds or a *Cleanup
// typed nil otherwise.
func FindCleanup(t has.Thing) has.Cleanup {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Cleanup); ok {
			return a
		}
	}
	return (*Cleanup)(nil)
}

// Found returns false if the receiver is nil otherwise true.
func (c *Cleanup) Found() bool {
	return c != nil
}

// Unmarshal is used to turn the passed data into a new Cleanup attribute.
func (*Cleanup) Unmarshal(data []byte) has.Attribute {
	c := NewCleanup(0, 0)
	for field, data := range decode.PairList(data) {
		data := []byte(data)
		switch field {
		case "AFTER":
			c.after = decode.Duration(data)
		case "JITTER":
			c.jitter = decode.Duration(data)
		default:
			log.Printf("Cleanup.unmarshal unknown attribute: %q: %q", field, data)
		}
	}
	return c
}

// Marshal returns a tag and []byte that represents the receiver.
func (c *Cleanup) Marshal() (tag string, data []byte) {
	tag = "cleanup"
	data = encode.PairList(
		map[string]string{
			"after":  string(encode.Duration(c.after)),
			"jitter": string(encode.Duration(c.jitter)),
		},
		'→',
	)
	return
}

func (c *Cleanup) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T After: %s Jitter: %s", c, c.after, c.jitter))
	buff = append(buff, DumpFmt("  %p %[1]T", c.Cancel))
	return
}

// Copy returns a copy of the Cleanup receiver. The copy will not inherit any
// pending clean up events.
func (c *Cleanup) Copy() has.Attribute {
	if c == nil {
		return (*Cleanup)(nil)
	}
	return NewCleanup(c.after, c.jitter)
}

// Cleanup schedules a clean up of the parent Thing. If there is already a
// clean up event pending it will be cancelled and a new one queued.
func (c *Cleanup) Cleanup() {
	if c == nil {
		return
	}

	// Cancel any outstanding clean up
	c.Abort()

	// If no parent Inventory have a clean up scheduled, schedule for reset. By
	// checking the parents we can put items into containers that have not been
	// moved yet and have the items cleaned up. If a container has been moved,
	// the item will be cleaned up when the parent container is cleaned up. For a
	// $CLEANUP the actor is where the clean up will take place.
	if !c.Active() {
		what := c.Parent()
		actor := FindLocate(what).Where().Parent()
		c.Cancel = event.Queue(actor, "$CLEANUP "+what.UID(), c.after, c.jitter)
	}
}

// Active returns true if any of the Inventories the parent Thing is in already
// have a clean up scheduled, otherwise false.
func (c *Cleanup) Active() bool {
	if c == nil {
		return false
	}

	if c.Cancel != nil {
		return true
	}

	if l := FindLocate(c.Parent()); l.Found() {
		if w := l.Where(); w != nil {
			if c := FindCleanup(w.Parent()); c.Found() {
				return c.Active()
			}
		}
	}

	return false
}

// Abort causes an outstanding clean up event to be cancelled for the parent
// Thing. If the Thing has an Inventory Abort is called on the contents
// recursively. If we don't do this putting an item into a container and then
// picking the container up would result in the item still being scheduled for
// a clean up and disappearing from the container.
func (c *Cleanup) Abort() {
	if c == nil {
		return
	}

	if p := c.Parent(); p != nil {
		if i := FindInventory(p); i.Found() {
			for _, t := range i.Contents() {
				if c := FindCleanup(t); c.Found() {
					c.Abort()
				}
			}
		}
	}

	if c.Cancel != nil {
		close(c.Cancel)
		c.Cancel = nil
	}
}

// Free makes sure references are nil'ed and channels closed when the Cleanup
// attribute is freed.
func (c *Cleanup) Free() {
	if c == nil {
		return
	}
	if c.Cancel != nil {
		close(c.Cancel)
		c.Cancel = nil
	}
	c.Attribute.Free()
}
