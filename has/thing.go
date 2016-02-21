// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

type Thing interface {
	Add(...Attribute)
	Attrs() []Attribute
	Dump() []string
	Remove(...Attribute)
}

type Attribute interface {
	Dump() []string
	Parent() Thing
	SetParent(Thing)
}
