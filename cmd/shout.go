// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: SHOUT <who> <message>
func init() {
	addHandler(shout{}, "SHOUT")
}

type shout cmd

func (shout) process(s *state) {
	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to shout something...")
		return
	}

	// Get all location inventories within 2 moves of current location, at a
	// distance of 1 location players will here what is shouted, at a distance of
	// 2 they will hear someone shout
	locations := attr.FindExits(s.where.Parent()).Within(2, s.where)

	// Try locking all of the locations we found
	lockAdded := false
	for _, d := range locations {
		for _, i := range d {
			if !s.CanLock(i) {
				s.AddLock(i)
				lockAdded = true
			}
		}
	}

	// If we added any locks return to the parser so we can relock
	if lockAdded {
		return
	}

	who := text.TitleFirst(attr.FindName(s.actor).TheName("Someone"))
	msg := strings.Join(s.input, " ")

	s.msg.Actor.SendGood("You shout: ", msg)
	s.msg.Observer.SendInfo(who, " shouts: ", msg)

	// Notify observers at a distance of 1
	s.msg.Observers.Filter(locations[1]...).SendInfo("You hear someone shout: ", msg)

	// Notify observers at a distance of 2
	s.msg.Observers.Filter(locations[2]...).SendInfo("You hear shouting nearby.")

	s.ok = true
	return
}
