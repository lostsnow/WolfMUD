// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Locate is used by any Thing that needs to know where it is. Locate has a
// reference to the Inventory that contains the parent Thing of this attribute.
// When using the default attr.Inventory type this reference is kept up to date
// automatically as the Thing is moved from Inventory to Inventory. Remember: A
// location in WolfMUD can be any Thing with an Inventory Attribute, not just
// conventional 'rooms'.
//
// Its default implementation is the attr.Locate type.
type Locate interface {
	Attribute

	// SetWhere is used to set the current Inventory.
	SetWhere(Inventory)

	// Where returns the Inventory currently set.
	Where() Inventory
}
