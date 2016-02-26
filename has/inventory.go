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

	// Add puts the specified Thing into the Inventory.
	Add(Thing)

	// Contents returns a []Thing representing the contents of the Inventory.
	Contents() []Thing

	// Crowded returns true if the Inventory is considered crowded otherwise
	// false. Definition of crowded is implementation dependant.
	Crowded() bool

	// Empty returns true if the Inventory is empty else false. What empty means
	// is up to the individual Inventory implementation. It may mean that the
	// Inventory is really empty or it may mean that there is nothing available
	// to be removed for example.
	Empty() bool

	// Found returns false if the receiver is nil otherwise true. A utility
	// method mainly for use with Finders such as attr.FindInventory.
	Found() bool

	// List returns a textual description of the Inventory content.
	List() string

	// LockID returns the unique locking ID for an Inventory.
	LockID() uint64

	// Remove takes the specified Thing out of the Inventory and returns the
	// Thing removed. Returns nil if Thing cannot be removed.
	Remove(Thing) Thing

	// Search returns the first Thing in an Inventory that has a matching Alias.
	// If there are no matches nil is returned.
	Search(alias string) Thing
}
