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

	l := attr.FindLocate(s.actor)
	to := l.Origin()

	if !s.CanLock(to) {
		s.AddLock(to)
		return
	}

	to.Add(s.actor)

	name := attr.FindName(s.actor).Name("something")
	s.msg.Observers[to].SendGood("There is a gentle pop and ", name, " appears.")
}
