// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Start is used to mark a Thing as being a starting location.
//
// Its default implementation is the attr.Start type.
type Start interface {
	Attribute

	// Pick returns an Inventory for a starting location. The location is picked
	// at random from all of the registered starting locations.
	Pick() Inventory
}
