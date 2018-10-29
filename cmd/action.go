// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"

	"strings"
)

// Syntax: $ACTION item
func init() {
	addHandler(action{}, "$action")
}

type action cmd

func (action) process(s *state) {

	l := len(s.locks)

	// Script the action
	c := strings.Join(s.input, " ")
	s.scriptAll(c)

	// If not relocking reschedule
	if l == len(s.locks) {
		attr.FindAction(s.actor).Action()
		s.ok = true
	}
}
