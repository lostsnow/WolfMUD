// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"strings"
)

func Inventory(t has.Thing) string {

	i := attr.FindInventory(t)

	if i == nil {
		return "You are not carrying anything."
	}

	buff := []string{}

	for _, i := range i.List() {
		if n := attr.FindName(i); n != nil {
			buff = append(buff, n.Name())
		}
	}

	if len(buff) == 0 {
		return "You are not carrying anything."
	}

	return "You are currently carrying:\n  " + strings.Join(buff, "\n  ")
}
