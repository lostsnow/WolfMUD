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

func Inventory(t has.Thing) (msg string, ok bool) {

	i := attr.Inventory().Find(t)

	if i == nil {
		msg = "You are not carrying anything."
		return
	}

	buff := []string{}

	for _, i := range i.List() {
		if n := attr.Name().Find(i); n != nil {
			buff = append(buff, n.Name())
		}
	}

	if len(buff) == 0 {
		msg = "You are not carrying anything."
		return
	}

	msg = "You are currently carrying:\n  " + strings.Join(buff, "\n  ")
	return msg, true
}
