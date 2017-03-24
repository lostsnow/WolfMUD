// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Attribute provides a minimal, common interface for implementing any type of
// Attribute. The interface provides a way for an Attribute to associate itself
// with the parent Thing it is being added to - or disassociate if removed.
// This allows any Attribute to access its parent Thing or other Attribute
// associated with the parent Thing.
//
// Its default implementation is the attr.Attribute type. For the different
// attributes available see the attr package.
//
// NOTE: the default implementation attr.Attribute does NOT provide a default
// Copy implementation. Each attribute must implement its own Copy method. This
// is due to the fact that other attributes will know best how to create copies
// based on their own implementation.
type Attribute interface {
	Dump() []string

	// Attributes need to be able to marshal and unmarshal themselves. Marshaler
	// has no default implementation and should be implemented by each Attribute.
	Marshaler

	// Found returns false if the receiver is nil otherwise true. Found has no
	// default implementation and should be implemented by each Attribute as it
	// is based on the receiver type.
	Found() bool

	// Parent returns the Thing to which the Attribute has been added.
	Parent() Thing

	// SetParent sets the Thing to which the Attribute has been added.
	SetParent(Thing)

	// Copy produces another, possibly inexact, instance of an Attribute. The
	// differences may be due to unique IDs, locks and other data that should not
	// be copied between instances.
	//
	// NOTE: The default implementation attr.Attribute does NOT implement Copy.
	// See main interface comments and attr.Attribute.
	Copy() Attribute

	// Free releases resources used by an attribute. It main use is to
	// disentangle cyclic pointers to assist in garbage collection. Attributes
	// that implement their own Free method should also call Attribute.Free.
	Free()
}
