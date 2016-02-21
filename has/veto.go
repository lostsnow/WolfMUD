// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

type Vetoes interface {
	Attribute
	Check(...string) Veto
}

type Veto interface {
	Command() string
	Dump() []string
	Message() string
}
