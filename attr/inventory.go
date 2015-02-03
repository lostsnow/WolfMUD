// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"strings"
)

type inventory struct {
	attribute
	contents []has.Thing
}

// Some interfaces we want to make sure we implement
var (
	_ has.Attribute = Inventory()
	_ has.Inventory = Inventory()
)

func Inventory() *inventory {
	return nil
}

func (*inventory) New(t ...has.Thing) *inventory {
	c := make([]has.Thing, len(t))
	copy(c, t)
	return &inventory{attribute{}, c}
}

func (*inventory) Find(t has.Thing) has.Inventory {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Inventory); ok {
			return a
		}
	}
	return nil
}

func (i *inventory) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T %d items:", i, len(i.contents)))
	for _, i := range i.contents {
		for _, i := range i.Dump() {
			buff = append(buff, DumpFmt("%s", i))
		}
	}
	return buff
}

func (i *inventory) Add(t has.Thing) {
	i.contents = append(i.contents, t)

	// Is what was added interested in where it is?
	if a := Locate().Find(t); a != nil {
		a.SetWhere(i.Parent())
	}
}

func (i *inventory) Remove(t has.Thing) has.Thing {
	for j, c := range i.contents {
		if c == t {
			// Is what was removed interested in where it is?
			if a := Locate().Find(t); a != nil {
				a.SetWhere(nil)
			}

			i.contents[j] = nil
			i.contents = append(i.contents[:j], i.contents[j+1:]...)
			return c
		}
	}
	return nil
}

func (i *inventory) Search(alias string) has.Thing {
	for _, c := range i.contents {
		if a := Alias().Find(c); a != nil {
			if a.HasAlias(alias) {
				return c
			}
		}
	}
	return nil
}

func (i *inventory) Contains(t has.Thing) bool {
	for _, c := range i.contents {
		if c == t {
			return true
		}
	}
	return false
}

func (i *inventory) List() []has.Thing {
	l := make([]has.Thing, len(i.contents))
	copy(l, i.contents)
	return l
}

func (i *inventory) Contents() string {
	buff := []string{}
	for _, c := range i.contents {
		if a := Name().Find(c); a != nil {
			buff = append(buff, a.Name())
		}
	}
	switch len(i.contents) {
	case 0:
		return "It is empty."
	case 1:
		return "It contains " + buff[0] + "."
	default:
		return "It contains:\n  " + strings.Join(buff, "\n  ")
	}
}
