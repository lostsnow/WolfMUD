// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type Narrative struct {
	*Inventory
}

// Some interfaces we want to make sure we implement
var (
	_ has.Inventory = &Narrative{}
	_ has.Narrative = &Narrative{}
)

func NewNarrative(t ...has.Thing) *Narrative {
	return &Narrative{NewInventory(t...)}
}

func FindNarrative(t has.Thing) has.Narrative {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Narrative); ok {
			return a
		}
	}
	return nil
}

func (n *Narrative) ImplementsNarrative() {}

func (n *Narrative) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T %d items:", n, len(n.contents)))
	for _, n := range n.contents {
		for _, i := range n.Dump() {
			buff = append(buff, DumpFmt("%s", i))
		}
	}
	return buff
}
