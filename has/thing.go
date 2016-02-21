// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

type Thing interface {
	Add(...Attribute)
	Remove(...Attribute)
	Attrs() []Attribute
	Dump() []string
}

type Attribute interface {
	Parent() Thing
	SetParent(Thing)
	Dump() []string
}
