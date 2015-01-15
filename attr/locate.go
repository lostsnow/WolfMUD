// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type locate struct {
	parent
	location has.Thing
}

func NewLocate(t has.Thing) *locate {
	l := &locate{parent{}, nil}
	if t != nil {
		l.SetLocation(t)
	}
	return l
}

func FindLocate(t has.Thing) has.Locate {

	compare := func(a has.Attribute) (ok bool) { _, ok = a.(has.Locate); return }

	if t := t.FindAttr(compare); t != nil {
		return t.(has.Locate)
	}
	return nil
}

func (l *locate) Dump() []string {
	name := ""
	if l := l.Location(); l != nil {
		if a := FindName(l); a != nil {
			name = a.Name()
		}
	}
	return []string{DumpFmt("%p %[1]T -> %p %s", l, l.location, name)}
}

func (l *locate) Location() has.Thing {
	return l.location
}

func (l *locate) SetLocation(to has.Thing) {
	l.location = to
}
