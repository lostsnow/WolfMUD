// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: SAY <message> | " <message>
func init() {
	addHandler(say{}, "SAY")
	addHandler(say{}, "\"")
}

type say cmd

func (say) process(s *state) {
	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("What did you want to say?")
		return
	}

	// Get all location inventories within 1 move of current location
	locations := attr.FindExits(s.where.Parent()).Within(1, s.where)

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

	// Is anyone else here? We can't call s.where.Occupied() as it will always
	// return true if s.actor is a Player.
	anybodyHere := false
	for _, t := range s.where.Contents() {
		if attr.FindPlayer(t).Found() && t != s.actor {
			anybodyHere = true
			break
		}
	}

	who := attr.FindName(s.actor).TheName("Someone")
	msg := strings.Join(s.input, " ")

	if !anybodyHere {
		s.msg.Actor.SendInfo("Talking to yourself again?")
	} else {
		s.msg.Actor.SendGood("You say: ", msg)
		s.msg.Observer.SendInfo(text.TitleFirst(who), " says: ", msg)
	}

	// Notify observers in near by locations
	s.msg.Observers.Filter(locations[1]...).SendInfo("You hear talking nearby.")

	s.ok = true
	return
}
