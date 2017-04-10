// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: $CLEANUP
func init() {
	AddHandler(Cleanup, "$cleanup")
}

func Cleanup(s *state) {

	// Find Inventory where clean up is going to take place and make sure we are
	// locking it
	to := attr.FindLocate(s.actor).Origin()
	if !s.CanLock(to) {
		s.AddLock(to)
		return
	}

	oc := attr.FindOnCleanup(s.actor)
	msg := oc.CleanupText()

	l := s.where
	p := l.Parent()
	e := attr.FindExits(p)

	// Clean up will not be seen if it does not happen in a location. It also
	// will not be seen if we have specifically have an empty message. So just
	// junk Thing.
	if !e.Found() || (oc.Found() && msg == "") {
		alias := s.actor.UID()
		s.scriptNone("junk", alias)
		s.ok = true
		return
	}

	// Clean up will be seen so add default message if we don't have one
	if !oc.Found() {
		name := attr.FindName(s.actor).Name("something")
		msg = "You are sure you noticed " + name + " here, but you can't see it now."
	}

	s.msg.Observers[l].SendInfo(msg)

	alias := s.actor.UID()
	s.scriptNone("junk", alias)
	s.ok = true
}
