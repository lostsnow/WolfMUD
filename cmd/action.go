// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: $ACTION item
func init() {
	AddHandler(Action, "$action")
}

func Action(s *state) {

	// Do we have item to perform action specified on command?
	if len(s.words) == 0 {
		return
	}

	// Search for item to perform action.
	alias := s.words[0]
	what := s.where.Search(alias)

	// If item not found all we can do is bail
	if what == nil {
		return
	}

	// Reschedule event and bail early if there are no players here to see the
	// action or it's too crowded to see the action.
	if !s.where.Players() || s.where.Crowded() {
		attr.FindAction(what).Action()
		s.ok = true
		return
	}

	// See if item actually has actions. If not, bail without rescheduling. There
	// is no point in rescheduling if there are no actions.
	oa := attr.FindOnAction(what)
	if !oa.Found() {
		return
	}

	// Display action and schedule next action. Only notify the actor if it's not
	// the thing issuing the command.
	if s.actor.UID() != what.UID() {
		s.msg.Actor.SendInfo(oa.ActionText())
	}
	s.msg.Observer.SendInfo(oa.ActionText())
	attr.FindAction(what).Action()

	s.ok = true
}
