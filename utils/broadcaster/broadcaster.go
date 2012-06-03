// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// BUG(Diddymus): This package is badly defined and needs reviewing. The
// Broadcast function should take a slice of responder.Interface and not
// thing.Interface. The AddThing method should not be here at all but is needed
// to add to the world which implements the broadcaster.Interface. Possibly
// broadcaster should be a sub package of the responder package?

// Package broadcaster defines the Interface for sending messages to multiple
// responders.
package broadcaster

import (
	"wolfmud.org/entities/thing"
)

type Interface interface {
	Broadcast(omit []thing.Interface, format string, any ...interface{})
	AddThing(thing thing.Interface)
}
