// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

import (
	"sync"
)

// Inventory are used to implement containers that can contain any type of
// Thing.
//
// Its default implementation is the attr.Inventory type.
type Inventory interface {
	Attribute
	sync.Locker

	// Contents returns a []Thing representing the contents of the Inventory.
	Contents() []Thing

	// Narratives returns a []Thing representing the narratives of the Inventory.
	Narratives() []Thing

	// Everything returns a []Thing representing the narratives and contents of
	// the inventory.
	Everything() []Thing

	// Crowded returns true if the Inventory is considered crowded otherwise
	// false. Definition of crowded is implementation dependant.
	Crowded() bool

	// Occupied returns true if there is at least one player in the Inventory.
	Occupied() bool

	// Empty returns true if the Inventory is empty else false. What empty means
	// is up to the individual Inventory implementation. It may mean that the
	// Inventory is really empty or it may mean that there is nothing available
	// to be removed for example.
	Empty() bool

	// List returns a textual description of the Inventory content.
	List() string

	// LockID returns the unique locking ID for an Inventory.
	LockID() uint64

	// Search returns the first Thing in an Inventory that has a matching Alias.
	// If there are no matches nil is returned.
	Search(alias string) Thing

	// Move removes a Thing from the receiver Inventory and places it into the
	// passed Inventory.
	Move(Thing, Inventory)

	// Carried return true if putting an item in an Inventory results in it being
	// carried by a player, otherwise false.
	Carried() bool

	// Outermost returns the top level inventory in an Inventory hierarchy.
	Outermost() Inventory

	// Disabled returns a slice of Thing for items that are out of play (disabled).
	Disabled() []Thing

	// Add puts a Thing into an Inventory and marks it as being initially out of
	// play (disabled).
	Add(Thing)

	// Remove takes a disabled Thing out of an Inventory.
	Remove(Thing)

	// Disable marks a Thing in an Inventory as being out of play.
	Disable(Thing)

	// Enable marks a Thing in an Inventory as being in play.
	Enable(Thing)
}
