// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Narrative is used to mark a Thing as being for narrative purposes. Any Thing
// can be a Narrative by adding a Narrative attribute.
//
// Its default implementation is the attr.Narrative type.
type Narrative interface {
	Attribute

	// ImplementsNarrative is a marker until we have a fuller implementation of
	// Narrative and we don't accidentally fulfil another interface.
	ImplementsNarrative()
}
