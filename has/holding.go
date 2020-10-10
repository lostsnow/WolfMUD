// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Holding represents the Things that a Thing is holding. That is the Things
// that are allocated to a Body attribute and have a usage of holding.
//
// The default implementation is the attr.Holding type.
type Holding interface {
	Attribute

	// Held returns a list of Thing references for Things that are currently
	// held.
	Held() []string
}
