// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: $RESET
func init() {
	AddHandler(Reset, "$reset")
	AddHandler(Reset, "$spawn")
}

func Reset(s *state) {

	// Find Inventory where reset is going to take place and make sure we are
	// locking it
	to := attr.FindLocate(s.actor).Origin()
	if !s.CanLock(to) {
		s.AddLock(to)
		return
	}

	or := attr.FindOnReset(s.actor)
	msg := or.ResetText()

	l, p := to, to.Parent()
	e := attr.FindExits(p)

	// Reset will not be seen if it does not happen in a location and we have no
	// message. It also will not be seen if we have specifically have an empty
	// message. So just add Thing.
	if (!e.Found() && !or.Found()) || (or.Found() && msg == "") {
		to.Add(s.actor)
		s.ok = true
		return
	}

	// Find out location where reset is happening. If reset is in a container we
	// need to know the location of the container. This is so that we know where
	// the reset will be seen so that we can send the reset message there.
	for !e.Found() {
		l = attr.FindLocate(p).Where()
		if l == nil {
			break
		}
		p = l.Parent()
		e = attr.FindExits(p)
	}

	// Make sure we are also locking the location where reset will be seen
	if l != nil && !s.CanLock(l) {
		s.AddLock(l)
		return
	}

	// Reset will be seen so add default message if we don't have one
	if !or.Found() {
		name := attr.FindName(s.actor).Name("something")
		msg = "You notice " + name + " that you didn't see before."
	}

	to.Add(s.actor)
	s.msg.Observers[l].SendInfo(msg)
	s.ok = true
}
