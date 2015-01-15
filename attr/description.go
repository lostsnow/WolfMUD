// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type description struct {
	parent
	description string
}

func NewDescription(d string) *description {
	return &description{parent{}, d}
}

func FindDescription(t has.Thing) has.Description {

	compare := func(a has.Attribute) (ok bool) { _, ok = a.(has.Description); return }

	if t := t.FindAttr(compare); t != nil {
		return t.(has.Description)
	}
	return nil
}

func (d *description) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", d, d.description)}
}

func (d *description) Description() string {
	return d.description
}
