// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Where(t has.Thing) (where has.Thing) {
	if t == nil {
		return nil
	}

	if l := attr.FindLocate(t); l != nil {
		return l.Location()
	}

	if where == nil && attr.FindExit(t) != nil {
		return t
	}

	return nil
}

func WhatWhere(alias string, t has.Thing) (what has.Thing, where has.Thing) {
	where = Where(t)

	if where != nil {
		if i := attr.FindInventory(where); i != nil {
			if what = i.Find(alias); what != nil {
				return what, where
			}
		}

		if n := attr.FindNarrative(where); n != nil {
			if what = n.Find(alias); what != nil {
				return what, where
			}
		}
	}

	if i := attr.FindInventory(t); i != nil {
		if what = i.Find(alias); what != nil {
			return what, t
		}
	}

	return nil, nil
}

func CheckVetoes(cmd string, what has.Thing) (string, bool) {
	if v := attr.FindVeto(what); v != nil {
		if v := v.Check(cmd); v != "" {
			return v, true
		}
	}

	return "", false
}
