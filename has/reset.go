// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Reset provides information on how often a Thing should reset and whether the
// Thing is respawned when picked up or not.
//
// The default implementation is the attr.Reset type.
type Reset interface {
	Attribute

	// Reset causes the parent Thing to be scheduled for a reset.
	Reset()

	// Resume a suspended Reset event.
	Resume()

	// Suspend a queued Reset event.
	Suspend()

	// Abort cancels any outstanding reset events.
	Abort()

	// Wait returns true if the Thing should wait for it's Inventory items to
	// reset before it resets, else false.
	Wait() bool

	// Spawn returns a non-spawnable copy of the parent Thing and schedules the
	// original to be respawned.
	Spawn() Thing

	// Spawnable returns true if the parent Thing is spawnable else false.
	Spawnable() bool

	// Spawned flags the Thing as being a spawned item.
	//
	// TODO(diddymus): This shouldn't be exposed in the interface and will be
	// removed in the attribute reorganisation.
	Spawned()

	// IsSpawned returns true if the Thing has been spawned else false.
	IsSpawned() bool

	// Unique returns true if item is considered unique else false.
	Unique() bool
}
