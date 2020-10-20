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

	var (
		what  has.Thing     // What we are resetting
		where has.Inventory // Container inventory we are resetting into
	)

	// Check container to reset into still valid, it may have been disposed of
	if s.actor.Freed() {
		return
	}

	// Find container's Inventory we will be resetting into, if not found bail
	if where = attr.FindInventory(s.actor); !where.Found() {
		return
	}

	// Make sure we are locking where reset will happen
	if !s.CanLock(where) {
		s.AddLock(where)
		return
	}

	// Find disabled Thing to be reset, if not found just return
	if what = where.SearchDisabled(s.words[0]); what == nil {
		return
	}

	r := attr.FindReset(what)

	// If we are resetting an original container (not spawned) should we wait for
	// its Inventory content to reset first? If so reschedule reset if content
	// not ready.
	if r.Found() && !r.IsSpawned() && r.Wait() {
		if i := attr.FindInventory(what); i.Found() && len(i.Disabled()) > 0 {
			r.Reset()
			return
		}
	}

	var (
		or    = attr.FindOnReset(what)
		msg   = or.ResetText()
		to, p = where, where.Parent()
		e     = attr.FindExits(p)
	)

	// Is the container we are resetting into disabled?
	whereDisabled := s.where != nil && s.where.SearchDisabled(s.actor.UID()) != nil

	// Reset will not be seen by players if:
	//
	//	- The container it happens in is disabled
	//	- The reset is not at a location and there is no specific reset message
	//	- If the reset message is disabled by a specific empty message
	//
	// If the reset is not seen by players then silently reset the Thing.
	if whereDisabled || (!e.Found() && !or.Found()) || (or.Found() && msg == "") {
		r.Abort()
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
