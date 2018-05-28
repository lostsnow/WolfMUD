// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Gender represents the gender of a Thing.
//
// Its default implementation is the attr.Gender type.
type Gender interface {
	Attribute

	// Gender returns a string representing a Thing's gender such as "Male",
	// "Female" or a non-specific "It".
	Gender() string
}
