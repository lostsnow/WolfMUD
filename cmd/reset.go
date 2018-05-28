// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: $RESET
//
// NOTE: The item will be 'out of play' so that we cannot find it by searching
// inventories for a passed alias. The only reference we have to it is the
// actor. This means that we cannot pass a unique alias to $RESET. As a
// consequence only a Thing can reset itself.
func init() {
	addHandler(reset{}, "$reset")
}

type reset cmd

func (reset) process(s *state) {

	// Find Inventory where reset is going to take place and make sure we are
	// locking it.
	//
	// TODO: Now we are using Inventory disabling is this check still required?
	origin := attr.FindLocate(s.actor).Origin()
	if !s.CanLock(origin) {
		s.AddLock(origin)
		return
	}

	or := attr.FindOnReset(s.actor)
	msg := or.ResetText()

	to, p := origin, origin.Parent()
	e := attr.FindExits(p)

	// Reset will not be seen if it does not happen in a location and we have no
	// message. The reset  will also not be seen if we specifically have an empty
	// message. In both cases just silently add the Thing.
	if (!e.Found() && !or.Found()) || (or.Found() && msg == "") {
		origin.Enable(s.actor)
		s.ok = true
		return
	}

	// Find out location where reset is happening. If reset is in a container we
	// need to know the location of the container. This is so that we know where
	// the reset will be seen so that we can send the reset message there.
	for !e.Found() {
		to = attr.FindLocate(p).Where()
		if to == nil {
			break
		}
		p = to.Parent()
		e = attr.FindExits(p)
	}

	// Make sure we are also locking the location where reset will be seen
	if to != nil && !s.CanLock(to) {
		s.AddLock(to)
		return
	}

	// Setup default message
	name := attr.FindName(s.actor).Name("something")
	def := "You notice " + name + " that you didn't see before."

	// Use default message if we don't have one
	if !or.Found() {
		msg = def
	}

	// Message will be seen if there are players at the reset location and the
	// location is not crowded.
	if to.Players() && !to.Crowded() {
		s.msg.Observers[to].SendInfo(msg)
	}

	// On the off chance that players may be in the container itself we send them
	// just the default message.
	if origin != to && origin.Players() && !origin.Crowded() {
		s.msg.Observers[origin].SendInfo(def)
	}

	attr.FindReset(s.actor).Abort()
	origin.Enable(s.actor)
	attr.FindAction(s.actor).Action()

	s.ok = true
}
