// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Wearing represents the Things that a Thing is wearing. That is the Things
// that are allocated to a Body attribute and have a usage of wearing.
//
// The default implementation is the attr.Wearing type.
type Wearing interface {
	Attribute

	// Held returns a list of Thing references for Things that are currently
	// held.
	Worn() []string
}
