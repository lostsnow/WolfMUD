// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: $CLEANUP
func init() {
	AddHandler(Cleanup, "$cleanup")
}

func Cleanup(s *state) {

	name := attr.FindName(s.actor).Name("something")
	s.msg.Observer.SendInfo(text.TitleFirst(name), " fades away and is gone.")

	alias := s.actor.UID()
	s.scriptNone("junk", alias)
	s.ok = true
}
