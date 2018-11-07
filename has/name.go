// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Name provides a short textual name for a Thing. Short names are usually of
// the form 'a bag', 'an apple', 'some rocks'.
//
// Its default implementation is the attr.Name type.
type Name interface {
	Attribute

	// Name returns the short name for a Thing. If the name cannot be returned
	// the preset can be used as a default.
	Name(preset string) string

	// TheName returns the short name for a Thing, but with any leading "A ",
	// "An " or "Some " changed to "The ".
	TheName(preset string) string
}
