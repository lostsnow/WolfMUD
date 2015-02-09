// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Read(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "Did you want to read something specific?"
		return
	}

	what := what(aliases[0], t)

	if what == nil {
		msg = "You see no '" + aliases[0] + "' to read."
		return
	}

	name := "something"
	if n := attr.Name().Find(what); n != nil {
		name = n.Name()
	}

	if w := attr.Writing().Find(what); w != nil {
		msg = "You read the writing on " + name + ". It says: " + w.Writing()
		return msg, true
	}

	msg = "You see no writing on " + name + " to read."
	return
}
