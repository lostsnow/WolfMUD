// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type Inventory struct {
	attribute
	contents []has.Thing
}

// Some interfaces we want to make sure we implement
var (
	_ has.Inventory = &Inventory{}
)

func NewInventory(t ...has.Thing) *Inventory {
	c := make([]has.Thing, len(t))
	copy(c, t)
	return &Inventory{attribute{}, c}
}

func FindInventory(t has.Thing) has.Inventory {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Inventory); ok {
			return a
		}
	}
	return nil
}

func (i *Inventory) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T %d items:", i, len(i.contents)))
	for _, i := range i.contents {
		for _, i := range i.Dump() {
			buff = append(buff, DumpFmt("%s", i))
		}
	}
	return buff
}

func (i *Inventory) Add(t has.Thing) {
	i.contents = append(i.contents, t)

	// Is what was added interested in where it is?
	if a := FindLocate(t); a != nil {
		a.SetWhere(i.Parent())
	}
}

func (i *Inventory) Remove(t has.Thing) has.Thing {
	for j, c := range i.contents {
		if c == t {
			// Is what was removed interested in where it is?
			if a := FindLocate(t); a != nil {
				a.SetWhere(nil)
			}

			i.contents[j] = nil
			i.contents = append(i.contents[:j], i.contents[j+1:]...)
			return c
		}
	}
	return nil
}

func (i *Inventory) Search(alias string) has.Thing {
	for _, c := range i.contents {
		if a := FindAlias(c); a != nil {
			if a.HasAlias(alias) {
				return c
			}
		}
	}
	return nil
}

func (i *Inventory) Contains(t has.Thing) bool {
	for _, c := range i.contents {
		if c == t {
			return true
		}
	}
	return false
}

func (i *Inventory) List() []has.Thing {
	l := make([]has.Thing, len(i.contents))
	copy(l, i.contents)
	return l
}

func (i *Inventory) Contents() string {
	buff := make([]byte, 0, 1024)

	switch len(i.contents) {
	case 0: // Empty? Just return
		return "It is empty."
	case 1: // Single item? Use a sentance: "It contains XXX."
		buff = append(buff, "It contains "...)
	default: // For multiple items display a list of them.
		buff = append(buff, "It contains:\n  "...)
	}

	mark := len(buff)

	for _, c := range i.contents {
		if a := Name().Find(c); a != nil {
			if len(buff) > mark {
				buff = append(buff, "\n  "...)
			}
			buff = append(buff, a.Name()...)
		}
	}

	// End single item sentance with a fullstop.
	if len(i.contents) == 1 {
		buff = append(buff, "."...)
	}

	return string(buff)
}
