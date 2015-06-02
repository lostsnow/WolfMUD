// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: ( LOOK | L )
//
// BUG(diddymus): If we look and where we are has a narrative attribute but no
// inventory attribute the narrative contents will be listed by mistake. Is
// this an example of a bigger problem with narratives/inventories still?
func Look(t has.Thing) (msg string, ok bool) {

	// Do we know where we are?
	var where has.Thing
	locater := attr.FindLocate(t)
	if locater != nil {
		where = locater.Where()
	}

	// Or are we the where?
	if locater == nil {
		if a := attr.FindInventory(t); a != nil {
			where = t
		}
	}

	// Still not anywhere?
	if where == nil {
		msg = "You are in a dark void. Around you nothing. No stars, no light, no heat and no sound."
		return
	}

	buff := make([]byte, 0, 1024)

	if a := attr.FindName(where); a != nil {
		buff = append(buff, "[ "...)
		buff = append(buff, a.Name()...)
		buff = append(buff, " ]\n"...)
	}

	mark := len(buff)

	for _, d := range attr.FindAllDescription(where) {
		buff = append(buff, d.Description()...)
		buff = append(buff, " "...)
	}

	// If we added descriptions chop off space appended to last description
	if mark != len(buff) {
		buff = buff[:len(buff)-1]
	}

	buff = append(buff, "\n\n"...)
	mark = len(buff)

	// Note: We don't want to include the looker in the list of things here which
	// is what the l != t check is for
	if a := attr.FindInventory(where); a != nil {
		for _, l := range a.Contents() {
			if l == t {
				continue
			}
			if n := attr.FindName(l); n != nil {
				buff = append(buff, "You can see "...)
				buff = append(buff, n.Name()...)
				buff = append(buff, " here.\n"...)
			}
		}
	}

	if mark != len(buff) {
		buff = append(buff, "\n"...)
	}

	if a := attr.FindExits(where); a != nil {
		buff = append(buff, a.List()...)
	} else {
		buff = append(buff, "You can see no immediate exits from here."...)
	}

	return string(buff), true
}
