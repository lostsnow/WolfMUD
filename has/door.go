// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Door provides a way of blocking travel in a specified direction when closed.
// A Door can also be used to implement door-like items such as a gate, a panel
// or a bookcase.
//
// Its default implementation is the attr.Door type.
type Door interface {
	Attribute

	// Open is used to change the state of a Door to open.
	Open()

	// Opened returns true if the state of a Door is open, otherwise false.
	Opened() bool

	// Close is used to change the state of a Door to closed.
	Close()

	// Closed returns true if the state of a Door is closed, otherwise false.
	Closed() bool

	// Direction returns the direction the door is blocking when closed. The
	// return values match the constants defined in attr.Exits.
	Direction() byte

	// OtherSide creates the opposing side of a Door.
	OtherSide()
}
