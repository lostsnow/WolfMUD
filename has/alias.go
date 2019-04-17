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

	// HasQualifier returns true if the qualifier passed is a valid qualifier,
	// otherwise false.
	HasQualifier(qualifier string) (found bool)

	// HasQualifierForAlias returns true if the qualifier passed is a valid
	// qualifier bound to the specified alias, otherwise false.
	HasQualifierForAlias(alias, qualifier string) (found bool)

	// Aliases returns all of the aliases as a []string or am empty slice if
	// there are no aliases.
	Aliases() []string

	// Qualifiers returns all of the qualifiers as a []string or am empty slice
	// if there are no qualifiers.
	Qualifiers() []string
}
