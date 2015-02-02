// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Read(t has.Thing, aliases []string) string {

	if len(aliases) == 0 {
		return "Did you want to read something specific?"
	}

	what, _ := WhatWhere(aliases[0], t)

	if what == nil {
		return "You see no '" + aliases[0] + "' to read."
	}

	name := "something"
	if n := attr.Name().Find(what); n != nil {
		name = n.Name()
	}

	if w := attr.FindWriting(what); w != nil {
		return "You read the writing on " + name + ". It says: " + w.Writing()
	}

	return "You see no writing on " + name + " to read."
}
