// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"fmt"
	"reflect"
)

type thing struct {
	attrs []has.Attribute
}

// Some interfaces we want to make sure we implement
var _ has.Thing = &thing{}

func Thing(a ...has.Attribute) has.Thing {
	t := &thing{}
	t.Add(a...)
	return t
}

func (t *thing) Add(a ...has.Attribute) {
	for _, a := range a {
		a.SetParent(t)
		t.attrs = append(t.attrs, a)
	}
}

func (t *thing) Remove(a ...has.Attribute) {
	for _, a := range a {
		for k, v := range t.attrs {
			if v == a {
				t.attrs[k] = nil
				a.SetParent(nil)
				t.attrs = append(t.attrs[:k], t.attrs[k+1:]...)
				break
			}
		}
	}
}

func (t *thing) Attrs() []has.Attribute {
	return t.attrs
}

// Find returns the first attribute matching the type of i or nil if there are
// no matches. Typically this function is only called from attribute finders
// such as the FindName function:
//
//	func FindName(t has.Thing) (n has.Name) {
//		n, _ = t.Find(&n).(has.Name)
//		return
//	}
//
func (t *thing) Find(i interface{}) interface{} {
	r := reflect.TypeOf(i).Elem()
	for _, a := range t.attrs {
		if reflect.TypeOf(a).Implements(r) {
			return reflect.ValueOf(a).Convert(r).Interface()
		}
	}
	return nil
}

func (t *thing) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T %d attributes:", t, len(t.attrs)))
	for _, a := range t.attrs {
		for _, a := range a.Dump() {
			buff = append(buff, DumpFmt("%s", a))
		}
	}
	return buff
}

func DumpFmt(format string, args ...interface{}) string {
	return "  " + fmt.Sprintf(format, args...)
}

type attribute struct {
	parent has.Thing
}

// Some interfaces we want to make sure we implement
// TODO: Is it odd attribute does not implement has.Attribute even though it is
// supposed to be the default implementation?
//var _ has.Attribute = &attribute{}

func (a *attribute) Parent() has.Thing {
	return a.parent
}

func (a *attribute) SetParent(t has.Thing) {
	a.parent = t
}
