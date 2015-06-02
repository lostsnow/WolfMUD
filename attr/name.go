// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/has"
)

type Name struct {
	Attribute
	name string
}

// Some interfaces we want to make sure we implement
var (
	_ has.Name = &Name{}
)

// Name implements an attribute for naming Things. It is used when referring to
// or listing Things. For example if there is a sword it could have the name of
// 'a sword'. Then manipulating it you could see the following messages:
//
//	You see a sword here.
//	You pick up a sword.
//	You examine a sword.
//	You start to wield a sword.
//
// Messages such as the examples would typically be general messages with a
// placeholder for the name of the Thing. For example:
//
//	You see %s here.
//	You pick up %s.
//	You examine %s.
//	You start to wield %s.
//
// It is therefore important to take this into consideration when choosing
// names for Things.
func NewName(n string) *Name {
	return &Name{Attribute{}, n}
}

// FindName searches the attributes of the specified Thing for attributes that
// implement has.Name returning all that match. If no matches are found an
// empty slice will be returned.
func FindName(t has.Thing) has.Name {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Name); ok {
			return a
		}
	}
	return nil
}

func (n *Name) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", n, n.name)}
}

// Name returns the name stored in the attribute.
func (n *Name) Name() string {
	return n.name
}
