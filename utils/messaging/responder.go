// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package responder implements a standard way to send responses to players.
package messaging

// Respond should be implemented by anything that wants to 'respond' to players.
// It is modelled after fmt.Printf so that messages can easily be built with
// parameters. For example:
//
//	cmd.Respond("You go %s.", directionLongNames[d])
type Responder interface {
	Respond(format string, any ...interface{})
}
