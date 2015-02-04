// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func WhatWhere(alias string, t has.Thing) (what has.Thing, where has.Thing) {

	// If thing locateable get where from there
	if a := attr.Locate().Find(t); a != nil {
		where = a.Where()
	}

	// If thing itself is exitable use that for where
	if where == nil {
		if attr.Exits().Find(t) != nil {
			where = t
		}
	}

	// If we know where we are check inventory and narratives
	if where != nil {
		if a := attr.Inventory().Find(where); a != nil {
			if what = a.Search(alias); what != nil {
				return what, where
			}
		}

		if a := attr.Narrative().Find(where); a != nil {
			if what = a.Search(alias); what != nil {
				return what, where
			}
		}
	}

	// If we haven't found our what and where yet check our thing's inventory
	if a := attr.Inventory().Find(t); a != nil {
		if what = a.Search(alias); what != nil {
			return what, t
		}
	}

	// Not found...
	return nil, nil
}

func CheckVetoes(cmd string, what has.Thing) has.Veto {
	if vetoes := attr.Vetoes().Find(what); vetoes != nil {
		if veto := vetoes.Check(cmd); veto != nil {
			return veto
		}
	}

	return nil
}
