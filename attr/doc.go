// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package attr (attributes) implements all of the functionality for objects
// in a WolfMUD world. All objects are instances of the Thing type. To these
// instances various attributes are added depending on the functionality
// required. The functionality of an object may be changed at runtime by adding
// or removing attributes.
//
// A Thing can be searched for specific Attribute types using finders. All
// finders return a typed nil in the event of the specific Attribute type not
// being found. Returning a typed nil such as (*Alias)(nil) allows the finder
// to be chained to any other methods taking the a pointer to the Attribute
// type as a receiver. For example:
//
//	if attr.FindAlias(t).HasAlias("test") {
//		// do something...
//	}
//
// If you need to check specifically if the finder returns nil, compare the
// returned value to the typed nil:
//
//	if a := attr.FindAlias(t); a == (*Alias)(nil) {
//		// do something...
//	}
//
// Or if available use the Found method:
//
//	if !attr.FindAlias(t).Found() {
//		// do something...
//	}
//
// All methods that take a pointer to an Attribute as a receiver are expected
// to be able to handle a nil receiver unless otherwise stated.
//
// All typed nils should be of the same type as the default implementation for
// the type. For example the has.Alias interface has attr.Alias as the default
// implementation, therefore the typed nil should also be of type *Alias.
//
// While it may seem unusual to return a typed nil it simplifies and reduces a
// lot of code in WolfMUD.
package attr
