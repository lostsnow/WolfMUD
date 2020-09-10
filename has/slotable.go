// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Slotable applies to any Thing that can be used (held, worn, wielded, etc.)
// in one or more Body slots.
//
// Slotable does not have a separate, default implementation and is not a
// standalone attribute that can be Marshaled/Unmarshaled like other
// attributes.
type Slotable interface {

	// Parent returns the Thing to which the Attribute has been added.
	Parent() Thing

	// Slots returns the Body slot references required to use a Thing. The
	// returned slice should not be modified, even temporally.
	Slots() []string
}
