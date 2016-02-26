// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Player is used to represent an actual player.
//
// Its default implementation is the attr.Player type.
type Player interface {
	Attribute

	// Found returns false if the receiver is nil otherwise true. A utility
	// method mainly for use with Finders such as attr.FindPlayer.
	Found() bool

	// Write implements the standard io.Writer interface. It is used to write
	// textual information back to the player.
	Write([]byte)
}
