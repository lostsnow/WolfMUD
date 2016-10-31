// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Description provides descriptive text for a Thing.
//
// Its default implementation is the attr.Description type.
type Description interface {
	Attribute

	// Description returns the descriptive text for the attribute.
	Description() string
}
