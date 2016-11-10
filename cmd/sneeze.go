// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: SNEEZE
func init() {
	AddHandler(Sneeze, "SNEEZE")
}

func Sneeze(s *state) {

	// Get all location inventories within 2 moves of current location
	locations := attr.FindExits(s.where.Parent()).Within(2)

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

	// Notify actor
	s.msg.Actor.Send("You sneeze. Aaahhhccchhhooo!")

	// Notify observers in same location
	who := attr.FindName(s.actor).Name("Someone")
	s.msg.Observer.Send("You see ", who, " sneeze.")

	// Notify observers in near by locations
	s.msg.Observers.Filter(locations[1]...).Send("You hear a loud sneeze.")

	// Notify observers in further out locations
	s.msg.Observers.Filter(locations[2]...).Send("You hear a sneeze.")

	s.ok = true
}
