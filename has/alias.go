// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Alias provides aliases that can be used to refer to a Thing.
//
// Its default implementation is the attr.Alias type.
type Alias interface {
	Attribute

	// HasAlias returns true if the alias passed is a valid alias, otherwise
	// false.
	HasAlias(alias string) (found bool)
}
