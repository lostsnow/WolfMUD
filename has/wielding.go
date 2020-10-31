// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Wielding represents the Things that a Thing is wielding. That is the Things
// that are allocated to a Body attribute and have a usage of wielding.
//
// The default implementation is the attr.Wielding type.
type Wielding interface {
	Attribute

	// Wielded returns a list of Thing references for Things that are currently
	// wielded.
	Wielded() []string
}
