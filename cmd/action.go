// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: $ACTION
func init() {
	AddHandler(Action, "$action")
}

func Action(s *state) {

	// Reschedule event and bail early if there are no players here to see the
	// action or it's too crowded to see the action.
	if !s.where.Players() || s.where.Crowded() {
		attr.FindAction(s.actor).Action()
		s.ok = true
		return
	}

	// See if item actually has actions. If not, bail without rescheduling. There
	// is no point in rescheduling if there are no actions.
	oa := attr.FindOnAction(s.actor)
	if !oa.Found() {
		return
	}

	// Display action and schedule next action
	s.msg.Observer.SendInfo(oa.ActionText())
	attr.FindAction(s.actor).Action()

	s.ok = true
}
