// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: ( INVENTORY | INV )
func init() {
	AddHandler(Inventory, "INV", "INVENTORY")
}

func Inventory(s *state) {

	// Try and find out if we are carrying anything
	inv := attr.FindInventory(s.actor).Contents()
	if len(inv) == 0 {
		s.msg.Actor.Send("You are not carrying anything.")
		return
	}

	s.msg.Actor.Send("You are currently carrying:")

	// List what we are carrying
	for _, i := range inv {
		s.msg.Actor.Send("  ", attr.FindName(i).Name("something"))
	}

	who := attr.FindName(s.actor).Name("Someone")
	s.msg.Observer.Send("You see ", who, " check over their gear.")

	s.ok = true
}
