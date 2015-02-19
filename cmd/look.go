// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Syntax: ( LOOK | L )
func Look(t has.Thing) (msg string, ok bool) {

	// Do we know where we are?
	var where has.Thing
	if a := attr.Locate().Find(t); a != nil {
		where = a.Where()
	}

	// Or are we the where?
	if where == nil {
		if a := attr.Inventory().Find(t); a != nil {
			where = t
		}
	}

	// Still not anywhere?
	if where == nil {
		msg = "You are in a dark void. Around you nothing. No stars, no light, no heat and no sound."
		return
	}

	buff := make([]byte, 0, 1024)
	attrName := attr.Name()

	if a := attrName.Find(where); a != nil {
		buff = append(buff, "[ "...)
		buff = append(buff, a.Name()...)
		buff = append(buff, " ]\n"...)
	}

	if a := attr.Description().Find(where); a != nil {
		buff = append(buff, a.Description()...)
	}

	buff = append(buff, "\n\n"...)
	mark := len(buff)

	// Note: We don't want to include the looker in the list of things here which
	// is what the l != t check is for
	if a := attr.Inventory().Find(where); a != nil {
		for _, l := range a.List() {
			if l == t {
				continue
			}
			if n := attrName.Find(l); n != nil {
				buff = append(buff, "You can see "...)
				buff = append(buff, n.Name()...)
				buff = append(buff, " here.\n"...)
			}
		}
	}

	if mark != len(buff) {
		buff = append(buff, "\n"...)
	}

	if a := attr.Exits().Find(where); a != nil {
		buff = append(buff, a.List()...)
	} else {
		buff = append(buff, "You can see no immediate exits from here."...)
	}

	return string(buff), true
}
