// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: ( INVENTORY | INV )
func init() {
	addHandler(inventory{}, "INV", "INVENTORY")
}

type inventory cmd

func (inventory) process(s *state) {

	// Try and find out if we are carrying anything
	inv := attr.FindInventory(s.actor).Contents()
	if len(inv) == 0 {
		s.msg.Actor.SendInfo("You are not carrying anything.")
		return
	}

	s.msg.Actor.Send("You currently have:")

	b := attr.FindBody(s.actor)

	// List what we are carrying
	for _, what := range inv {
		s.msg.Actor.Send("  ", attr.FindName(what).Name("something"))
		if b.Using(what) {
			s.msg.Actor.Append(" - ", text.Green, b.Usage(what), text.Reset)
		}
	}

	who := attr.FindName(s.actor).Name("Someone")
	s.msg.Observer.SendInfo("You see ", who, " check over their gear.")

	s.ok = true
}
