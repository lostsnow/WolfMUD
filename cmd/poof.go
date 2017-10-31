// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: $POOF
func init() {
	addHandler(poof{}, "$POOF")
}

type poof cmd

func (poof) process(s *state) {

	name := attr.FindName(s.actor).Name("Someone")

	s.msg.Observer.SendGood("There is a cloud of smoke from which ", name, " emerges coughing and spluttering.")

	s.scriptActor("LOOK")
}
