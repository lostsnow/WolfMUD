// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Holdable represents the Body slots required to be able to hold a Thing. Any
// Thing with a Holdable attribute can be held providing the specified Body
// slots are available.
//
// Its default implementation is the attr.Holdable type.
type Holdable interface {
	Attribute
	Slotable

	// Method to distinguish between Holdable, Wearable and Wieldable interfaces.
	IsHoldable() bool
}
