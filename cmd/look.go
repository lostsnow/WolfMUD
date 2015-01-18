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

// Look can handle something with exits or something that is located somewhere.
func Look(t has.Thing) string {

	where := Where(t)
	if where == nil {
		return "You are in a dark void. Around you nothing. No stars, no light, no heat and no sound."
	}

	buff := []string{}

	if a := attr.FindName(where); a != nil {
		buff = append(buff, "[ "+a.Name()+" ]")
	}

	if a := attr.FindDescription(where); a != nil {
		buff = append(buff, a.Description())
	}

	buff = append(buff, "")
	mark := len(buff)

	// Note: We don't want to include the looker in the list of things here which
	// is what the l != t check is for
	if a := attr.FindInventory(where); a != nil {
		for _, l := range a.List() {
			if n := attr.FindName(l); l != t && n != nil {
				buff = append(buff, "You can see "+n.Name()+" here.")
			}
		}
	}

	if mark != len(buff) {
		buff = append(buff, "")
	}

	if a := attr.FindExit(where); a != nil {
		buff = append(buff, a.List())
	} else {
		buff = append(buff, "You can see no immediate exits from here.")
	}

	return strings.Join(buff, "\n")
}
