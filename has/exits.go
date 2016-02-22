// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

type Exits interface {
	Attribute
	AutoLink(byte, Inventory)
	AutoUnlink(byte)
	ToName(direction byte) string
	Found() bool
	LeadsTo(string) Inventory
	Link(byte, Inventory)
	List() string
	NormalizeDirection(string) string
	Surrounding() []Inventory
	Unlink(byte)
	Within(int) [][]Inventory
}
