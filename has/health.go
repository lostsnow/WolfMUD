// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Health represents the state of health of a Thing.
//
// Its default implementation is the attr.Health type.
type Health interface {
	Attribute

	// State returns the current and maximum health points.
	State() (current, maximum int)

	// Adjust increases or decreses the current health points by the given amount.
	// The new value will be a minimum of 0 and capped at the health maximum.
	Adjust(int)

	// AutoUpdate enables or disables the automatic regeneration of current
	// health points.
	AutoUpdate(bool)

	// Prompt returns the current health if brief is true else current health and
	// maximum health as 'current/maximum' if false.
	Prompt(brief bool) []byte
}
