// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Syntax: ( INVENTORY | INV )
func Inventory(t has.Thing) (msg string, ok bool) {

	// Try and find our inventory
	i := attr.FindInventory(t)
	if i == nil {
		msg = "You are not carrying anything."
		return
	}

	buff := make([]byte, 0, 1024)

	for _, i := range i.List() {
		if n := attr.FindName(i); n != nil {
			buff = append(buff, "\n  "...)
			buff = append(buff, n.Name()...)
		}
	}

	if len(buff) == 0 {
		msg = "You are not carrying anything."
		return
	}

	msg = "You are currently carrying:" + string(buff)
	return msg, true
}
