// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type description struct {
	attribute
	description string
}

// Some interfaces we want to make sure we implement
var _ has.Attribute = &description{}
var _ has.Description = &description{}

func NewDescription(d string) *description {
	return &description{attribute{}, d}
}

func FindDescription(t has.Thing) (d has.Description) {
	d, _ = t.Find(&d).(has.Description)
	return
}

func (d *description) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", d, d.description)}
}

func (d *description) Description() string {
	return d.description
}
