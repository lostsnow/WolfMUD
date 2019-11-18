// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/message"
)

// Syntax: $RESET <unique-alias>
//
// For the $RESET command the actor should be the Thing with the Inventory
// where the disabled Thing with the unique ID to reset is located.
func init() {
	addHandler(reset{}, "$reset")
}

type reset cmd

func (reset) process(s *state) {

	// The reference to the Actor may be stale and already freed due to event
	// queuing. This can occur with nested containers where the parent gets
	// cleaned up or reset. If actor is already freed just return.
	if s.actor.Freed() {
		return
	}

	// Find where reset will happen. We cannot use s.where as we want the actor's
	// inventory, not where the actor is. If we can't find where, maybe we are in
	// a container that has already been reset, all we can do is exit.
	where := attr.FindInventory(s.actor)
	if !where.Found() {
		return
	}

	// Make sure we are locking where reset will happen
	if !s.CanLock(where) {
		s.AddLock(where)
		return
	}

	// Find disabled Thing to be reset, if not found just return
	var what has.Thing
	for _, t := range where.Disabled() {
		if t.UID() == s.words[0] {
			what = t
			break
		}
	}
	if what == nil {
		return
	}

	or := attr.FindOnReset(what)
	msg := or.ResetText()

	to, p := where, where.Parent()
	e := attr.FindExits(p)

	// Reset will not be seen if it does not happen in a location and we have no
	// message. The reset will also not be seen if we specifically have an empty
	// message. In both cases just silently reset the Thing.
	if (!e.Found() && !or.Found()) || (or.Found() && msg == "") {
		where.Enable(what)
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
	name := attr.FindName(what).Name("something")
	def := "You notice " + name + " that you didn't see before."

	// Use default message if we don't have one
	if !or.Found() {
		msg = def
	}

	// Message will be seen if there are players at the reset location and the
	// location is not crowded.
	if to.Occupied() && !to.Crowded() {
		s.msg.Observers[to].SendInfo(msg)
	}

	// On the off chance that players may be in the container itself we send them
	// just the default message. We need to manually allocate a buffer as we only
	// get buffers for the outermost inventories automatically when locking.
	if where != to && where.Occupied() && !where.Crowded() {
		s.msg.Observers[where] = message.AcquireBuffer()
		s.msg.Observers[where].SendInfo(def)
	}

	attr.FindReset(what).Abort()
	where.Enable(what)
	attr.FindAction(what).Action()

	s.ok = true
}
