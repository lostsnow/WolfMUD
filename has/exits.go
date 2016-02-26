// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Exits coordinate linkages and movement between location Inventory
// attributes.
//
// Its default implementation is the attr.Exits type.
type Exits interface {
	Attribute

	// AutoLink links two opposing exits. Autolink links the passed Inventory to
	// the receiver's exit for the given direction. It then links the passed
	// Inventory's Exits to the receiver's Inventory in the opposite direction.
	AutoLink(direction byte, to Inventory)

	// AutoUnlink unlinks two opposing exits. Autounlink unlinks the receiver's
	// exit for the given direction. It then unlinks the passed Inventory's Exits
	// in the opposite direction.
	AutoUnlink(direction byte)

	// LeadsTo returns the Inventory of the location reached by taking the exit
	// in the given direction. If the exit does not leads nowhere nil returned.
	LeadsTo(direction byte) Inventory

	// Link links the receiver's exit for the given direction to the passed
	// Inventory.
	Link(direction byte, to Inventory)

	// List returns a string describing the exits available for the receiver.
	List() string

	// NormalizeDirection takes a direction name such as 'North', 'north',
	// 'NoRtH' or 'N' and returns the direction's index. If the name cannot be
	// normalized a non-nil error will be returned.
	NormalizeDirection(name string) (byte, error)

	// Surrounding returns a slice of Inventory, one Inventory for each location
	// reachable via the receiver's Exits. If no locations are reachable an empty
	// slice is returned.
	Surrounding() []Inventory

	// ToName takes a direction index and returns the long lowercased name such
	// as 'north' or 'northwest'.
	ToName(direction byte) string

	// Unlink unlinks the Inventory from the receiver for the given direction.
	Unlink(direction byte)

	// Within returns all location Inventories within the given number of moves
	// from the receiver's location. The inventories are returned as a slice of
	// Inventory slices. The first slice is the number of moves from the current
	// location. The second slice is a list of the Inventory reachable for that
	// number of moves.
	Within(moves int) [][]Inventory
}
