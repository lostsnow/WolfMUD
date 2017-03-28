// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Thing is used to create everything and anything in a WolfMUD world. In
// WolfMUD everything is created by creating a Thing and adding Attributes to
// it. Attribute define the behaviour and characteristics of specific Things.
// Attributes may be added and removed at runtime to dynamically affect a
// Thing.
//
// Its default implementation is the attr.Thing type. For the different
// attributes available see the attr package.
type Thing interface {

	// Add is used to add one or more Attribute to a Thing.
	Add(...Attribute)

	// Attrs returns a []Attribute of all the Attribute for a Thing.
	Attrs() []Attribute

	Dump() []string

	// Remove is used to remove one or more Attribute from a Thing.
	Remove(...Attribute)

	// Free is used to clean-up/release references to all Attribute for a Thing.
	Free()

	// Copy produces another, possibly inexact, instance of a Thing. The
	// differences may be due to unique IDs, locks and other data that should not
	// be copied between instances. The copy will contain a copy of all of the
	// attributes and possibly other Things associated with the Thing as well.
	Copy() Thing

	// SetOrigins updates the origin for the Thing to its containing Inventory and
	// recursivly sets the origins for the content of a Thing's Inventory if it has
	// one.
	SetOrigins()

	// Dispose will either reset or discard a Thing when it is finished with.
	Dispose()

	// UID returns the unique identifier for a Thing or an empty string if the
	// unique ID is unavailable.
	UID() string
}
