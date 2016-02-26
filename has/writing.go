// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Writing allows text to be added to any Thing that can then be read to reveal
// what is written.
//
// Its default implementation is the attr.Writing type.
type Writing interface {
	Attribute

	// Writing returns the text for what has been written.
	Writing() string
}
