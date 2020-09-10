// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Wieldable represents the Body slots required to be able to wield a Thing.
// Any Thing with a Wieldable attribute is wieldable providing the specified
// Body slots are available.
//
// Its default implementation is the attr.Wieldable type.
type Wieldable interface {
	Attribute
	Slotable

	// Method to distinguish between Holdable, Wearable and Wieldable interfaces.
	IsWieldable() bool
}
