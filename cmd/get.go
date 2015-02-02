// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Get(t has.Thing, aliases []string) string {

	if len(aliases) == 0 {
		return "You go to get... something?"
	}

	what, where := WhatWhere(aliases[0], t)

	if where == t || what == nil {
		return "You see no '" + aliases[0] + "' to get."
	}

	to := attr.Inventory().Find(t)
	name := attr.Name().Find(what).Name()

	if veto := CheckVetoes("GET", what); veto != nil {
		return veto.Message()
	}

	if attr.Inventory().Find(where).Remove(what) == nil {
		return "You cannot get " + name + "."
	}

	to.Add(what)

	return "You get " + name + "."
}
