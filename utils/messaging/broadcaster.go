// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// BUG(Diddymus): This package is badly defined and needs reviewing. The
// Broadcast function should take a slice of responder.Interface and not
// thing.Interface.

package messaging

import (
	"code.wolfmud.org/WolfMUD.git/entities/thing"
)

// Broadcast should be implemented by anything that wants to send messages to
// multiple responders. This is usually to everyone currently in the world
// or at a specific location. Like responders the function is modelled after
// fmt.Printf and takes messages formatted in the same way. The omit parameter
// is used to omit certain responders. For example if a player sneezes in a
// location they would have a different message and be omitted from the broadcast
// to the location and the sneezer and people in the location would be omitted
// from the message broadcast to the world:
//
//	cmd.Respond("You sneeze. Aaahhhccchhhooo!")
//	cmd.Broadcast([]thing.Interface{p}, "You see %s sneeze.", cmd.Issuer.Name())
//	PlayerList.Broadcast(p.Locate().List(), "You hear a loud sneeze.")
//
type Broadcaster interface {
	Broadcast(omit []thing.Interface, format string, any ...interface{})
}
