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

	// Spawn returns a non-spawnable copy of the parent Thing and schedules the
	// original to be respawned.
	Spawn() Thing
}
