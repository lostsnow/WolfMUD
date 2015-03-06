// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type Name struct {
	attribute
	name string
}

// Some interfaces we want to make sure we implement
var (
	_ has.Name = &Name{}
)

func NewName(n string) *Name {
	return &Name{attribute{}, n}
}

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

func (n *Name) Name() string {
	return n.name
}
