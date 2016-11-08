// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"

	"strings"
)

// Syntax: SAY <message> | " <message>
func init() {
	AddHandler(Say, "SAY")
	AddHandler(Say, "\"")
}

func Say(s *state) {
	if len(s.words) == 0 {
		s.msg.Actor.WriteStrings("What did you want to say?")
		return
	}

	// Are we somewhere?
	if s.where == nil {
		s.msg.Actor.WriteStrings("There is nobody here to talk to.")
		return
	}

	// Is anyone else here?
	anybodyHere := false
	for _, t := range s.where.Contents() {
		if attr.FindPlayer(t).Found() && t != s.actor {
			anybodyHere = true
			break
		}
	}
	if !anybodyHere {
		s.msg.Actor.WriteStrings("Talking to yourself again?")
		return
	}

	// Get all location inventories within 1 move of current location
	locations := attr.FindExits(s.where.Parent()).Within(1)

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

	who := attr.FindName(s.actor).Name("Someone")
	msg := strings.Join(s.input, " ")

	s.msg.Actor.WriteStrings("You say: ", msg)
	s.msg.Observer.WriteStrings(who, " says: ", msg)

	// Notify observers in near by locations
	for _, e := range locations[1] {
		s.msg.Observers[e].WriteStrings("You hear talking nearby.")
	}

	s.ok = true
	return
}
