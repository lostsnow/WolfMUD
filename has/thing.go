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
}

// Attribute provides a minimal, common interface for implementing any type of
// Attribute. The interface provides a way for an Attribute to associate itself
// with the parent Thing it is being added to - or disassociate if removed.
// This allows any Attribute to access its parent Thing or other Attribute
// associated with the parent Thing.
//
// Its default implementation is the attr.Attribute type. For the different
// attributes available see the attr package.
type Attribute interface {
	Dump() []string

	// Found returns false if the receiver is nil otherwise true. Found has no
	// default implementation and should be implemented by each Attribute as it
	// is based on the receiver type.
	Found() bool

	// Parent returns the Thing to which the Attribute has been added.
	Parent() Thing

	// SetParent sets the Thing to which the Attribute has been added.
	SetParent(Thing)
}
