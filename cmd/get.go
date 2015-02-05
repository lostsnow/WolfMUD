// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Get(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "You go to get... something?"
		return
	}

	what, where := WhatWhere(aliases[0], t)

	if where == t || what == nil {
		msg = "You see no '" + aliases[0] + "' to get."
		return
	}

	to := attr.Inventory().Find(t)
	name := attr.Name().Find(what).Name()

	if veto := CheckVetoes("GET", what); veto != nil {
		msg = veto.Message()
		return
	}

	if attr.Inventory().Find(where).Remove(what) == nil {
		msg = "You cannot get " + name + "."
		return
	}

	to.Add(what)

	msg = "You get " + name + "."
	return msg, true
}
