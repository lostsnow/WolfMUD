// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type locate struct {
	attribute
	where has.Thing
}

// Some interfaces we want to make sure we implement
var _ has.Attribute = &locate{}
var _ has.Locate = &locate{}

func NewLocate(t has.Thing) *locate {
	l := &locate{attribute{}, nil}
	l.SetWhere(t)
	return l
}

func FindLocate(t has.Thing) (l has.Locate) {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Locate); ok {
			return a
		}
	}
	return nil
}

func (l *locate) Dump() []string {
	name := ""
	if w := l.Where(); w != nil {
		if a := FindName(w); a != nil {
			name = a.Name()
		}
	}
	return []string{DumpFmt("%p %[1]T -> %p %s", l, l.where, name)}
}

func (l *locate) Where() has.Thing {
	return l.where
}

// TODO: Should we be checking that w has an inventory if we are being placed
// there?
func (l *locate) SetWhere(w has.Thing) {
	l.where = w
}
