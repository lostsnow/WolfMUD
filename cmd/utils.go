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

	if w := attr.Locate().Find(t); w != nil {
		return w.Where()
	}

	if where == nil && attr.Exit().Find(t) != nil {
		return t
	}

	return nil
}

func WhatWhere(alias string, t has.Thing) (what has.Thing, where has.Thing) {
	where = Where(t)

	if where != nil {
		if i := attr.Inventory().Find(where); i != nil {
			if what = i.Search(alias); what != nil {
				return what, where
			}
		}

		if n := attr.FindNarrative(where); n != nil {
			if what = n.Search(alias); what != nil {
				return what, where
			}
		}
	}

	if i := attr.Inventory().Find(t); i != nil {
		if what = i.Search(alias); what != nil {
			return what, t
		}
	}

	return nil, nil
}

func CheckVetoes(cmd string, what has.Thing) has.Veto {
	if vetoes := attr.FindVeto(what); vetoes != nil {
		if veto := vetoes.Check(cmd); veto != nil {
			return veto
		}
	}

	return nil
}
