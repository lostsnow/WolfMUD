// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Move(t has.Thing, cmd string) (msg string, ok bool) {

	// A thing can only move itself if it knows where it is
	from := attr.Locate().Find(t)
	if from == nil || from.Where() == nil {
		msg = "You can't go anywhere. You don't know where you are!"
		return
	}

	// Is where we are exitable?
	exits := attr.Exits().Find(from.Where())
	if exits == nil {
		msg = "You can't see anywhere to go from here."
		return
	}

	if msg, ok = exits.Move(t, cmd); !ok {
		return
	}

	// Describe where we now are
	return Parse(t, "LOOK")
}
