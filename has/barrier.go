// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Barrier provides a way of conditionally vetoing movement in a given
// direction based on aliases. The barrier may be invisible or invisible and
// interactive or non-interactive depending on the Thing the attribute is
// attached to.
type Barrier interface {
	Attribute

	// Allowed returns a slice of aliases allowed to pass through the barrier.
	Allowed() []string

	// Denied returns a slice of aliases not allowed to pass through the barrier.
	Denied() []string
}
