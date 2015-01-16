// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type name struct {
	parent
	name string
}

func NewName(n string) *name {
	return &name{parent{}, n}
}

func (n *name) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", n, n.name)}
}

func FindName(t has.Thing) has.Name {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Name); ok {
			return a
		}
	}
	return nil
}

func (n *name) Name() string {
	return n.name
}
