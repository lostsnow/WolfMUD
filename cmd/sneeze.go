// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: SNEEZE
func init() {
	AddHandler(Sneeze, "SNEEZE")
}

func Sneeze(s *state) {

	var locations [][]has.Inventory

	// Incrementally get all locations within a radius of 2 moves from current
	// location, re-locking as we expand the radius so that FindExits is always
	// called on locked locations.
	for radius := 1; radius < 3; radius++ {

		// Get all location Inventory within current radius
		locations = attr.FindExits(s.where.Parent()).Within(radius)

		// Try locking all of the locations we found
		lockAdded := false
		for _, d := range locations {
			for _, i := range d {
				if !s.CanLock(i) {
					s.AddLock(i)
					lockAdded = true
				}
			}
			// If we added any locks return to the parser so we can relock
			if lockAdded {
				return
			}
		}

	}

	// Notify actor
	s.msg.Actor.SendGood("You sneeze. Aaahhhccchhhooo!")

	// Notify observers in same location
	who := attr.FindName(s.actor).Name("Someone")
	s.msg.Observer.SendInfo("You see ", who, " sneeze.")

	// Notify observers in near by locations
	s.msg.Observers.Filter(locations[1]...).SendInfo("You hear a loud sneeze.")

	// Notify observers in further out locations
	s.msg.Observers.Filter(locations[2]...).SendInfo("You hear a sneeze.")

	s.ok = true
}
