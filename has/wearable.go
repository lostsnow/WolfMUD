// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Wearable represents the Body slots required to be able to wear a Thing. Any
// Thing with a Wearable attribute is wearable providing the specified Body
// slots are available.
//
// Its default implementation is the attr.Wearable type.
type Wearable interface {
	Attribute
	Slotable

	// Method to distinguish between Holdable, Wearable and Wieldable interfaces.
	IsWearable() bool
}
