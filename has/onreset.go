// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// OnReset provides a reset or respawn message for a Thing.
//
// Its default implementation is the attr.OnReset type.
type OnReset interface {
	Attribute

	// ResetText returns the reset or respawn message for a Thing.
	ResetText() string
}
