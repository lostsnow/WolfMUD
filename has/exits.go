// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

type Exits interface {
	Attribute
	Link(byte, Inventory)
	AutoLink(byte, Inventory)
	Unlink(byte)
	AutoUnlink(byte)
	List() string
	NormalizeDirection(string) string
	LeadsTo(string) Inventory
	Found() bool
	Within(int) [][]Inventory
	Surrounding() []Inventory
}
