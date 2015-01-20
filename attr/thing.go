// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"fmt"
)

type thing struct {
	a []has.Attribute
}

// Some interfaces we want to make sure we implement
// TODO: Should thing implement the has.Attribute interface? Would it simplify
// things if it did?
var _ has.Thing = &thing{}

func Thing(a ...has.Attribute) has.Thing {
	t := &thing{}
	t.Add(a...)
	return t
}

func (t *thing) Add(a ...has.Attribute) {
	for _, a := range a {
		a.SetParent(t)
		t.a = append(t.a, a)
	}
}

func (t *thing) Remove(a ...has.Attribute) {
	for _, a := range a {
		for k, v := range t.a {
			if v == a {
				t.a[k] = nil
				t.a = append(t.a[:k], t.a[k+1:]...)
			}
		}
	}
}

func (t *thing) Attrs() []has.Attribute {
	return t.a
}

func (t *thing) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T %d attributes:", t, len(t.a)))
	for _, a := range t.a {
		for _, a := range a.Dump() {
			buff = append(buff, DumpFmt("%s", a))
		}
	}
	return buff
}

func DumpFmt(format string, args ...interface{}) string {
	return "  " + fmt.Sprintf(format, args...)
}

type parent struct {
	p has.Thing
}

// Some interfaces we want to make sure we implement
// TODO: Is it odd parent does not implement has.Attribute even though it is
// supposed to be the default implementation?
//var _ has.Attribute = &parent{}

func (p *parent) Parent() has.Thing {
	return p.p
}

func (p *parent) SetParent(t has.Thing) {
	p.p = t
}
