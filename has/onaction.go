// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// OnAction provides action messages for a Thing.
//
// Its default implementation is the attr.OnAction type.
type OnAction interface {
	Attribute

	// ActionText returns an action message for a Thing.
	ActionText() string
}
