// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Action provides information on how often a Thing should emit action
// messages.
//
// The default implementation is the attr.Action type.
type Action interface {
	Attribute

	// Action causes the parent Thing to schedule an action message.
	Action()

	// Reschedule an Action event based on the time the event was expected to
	// fire.
	Reschedule()

	// Abort cancels any outstanding action events.
	Abort()
}
