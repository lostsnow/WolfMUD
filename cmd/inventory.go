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

	// Try and find our inventory
	i := attr.FindInventory(s.actor)
	if i == nil {
		s.msg.actor.WriteString("You can't carry anything!")
		return
	}

	// Remember where we are in the buffer in case we want to rewind the next
	// write in the case of not actually carrying anything...
	rewind := s.msg.actor.Len()
	s.msg.actor.WriteString("You are currently carrying:")

	// Mark where we are in the buffer so we can check if we write any new data into it
	mark := s.msg.actor.Len()

	for _, i := range i.Contents() {
		if n := attr.FindName(i); n != nil {
			s.msg.actor.WriteJoin("\n  ", n.Name())
		}
	}

	// If no new data written to the buffer since 'mark', rewind it and write new message
	if mark == s.msg.actor.Len() {
		s.msg.actor.Truncate(rewind)
		s.msg.actor.WriteString("You are not carrying anything.")
	}

	s.ok = true
}
