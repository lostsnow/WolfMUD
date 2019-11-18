// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: $CLEANUP <unique-alias>
//
// For the $CLEANUP command the actor should be the Thing with the Inventory
// where the Thing with the unique ID to be cleaned up is located.
func init() {
	addHandler(cleanup{}, "$cleanup")
}

type cleanup cmd

func (cleanup) process(s *state) {

	// The reference to the Actor may be stale and already freed due to event
	// queuing. This can occur with nested containers where the parent gets
	// cleaned up or reset. If actor is already freed just return.
	if s.actor.Freed() {
		return
	}

	// Find where clean up will happen. We cannot use s.where as we want the
	// actor's inventory, not where the actor is. If we can't find where, maybe
	// we are in a container that has already been cleaned up, all we can do is
	// exit.
	where := attr.FindInventory(s.actor)
	if !where.Found() {
		return
	}

	// Make sure we are locking where clean up will happen
	if !s.CanLock(where) {
		s.AddLock(where)
		return
	}

	// Find thing to be cleaned up, if not found just return
	what := where.Search(s.words[0])
	if what == nil {
		return
	}

	oc := attr.FindOnCleanup(what)
	msg := oc.CleanupText()

	// Clean up will not be seen if we specifically have an empty message. It
	// will also not be seen if there are no players here to see it or it's too
	// crowded. In these cases just junk Thing.
	if (oc.Found() && msg == "") || !where.Occupied() || where.Crowded() {
		s.scriptNone("junk", what.UID())

		// s.ok will be that of the scripted JUNK and we will also retry if JUNK
		// modifies the locks
		return
	}

	// Clean up will be seen so add default message if we don't have one
	if !oc.Found() {
		name := attr.FindName(what).Name("something")
		msg = "You are sure you noticed " + name + " here, but you can't see it now."
	}

	// Script JUNK of item to remove it and relock/abort if needed
	l := len(s.locks)
	s.scriptNone("junk", what.UID())
	if l != len(s.locks) || !s.ok {
		return
	}

	// Display message where clean up happens
	s.msg.Observers[where].SendInfo(msg)

	s.ok = true
}
