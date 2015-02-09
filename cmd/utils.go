// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func what(alias string, t has.Thing) (what has.Thing) {

	// If thing knows it's in an inventory try that inventory first
	if a := attr.Locate().Find(t); a != nil {
		if what = search(alias, a.Where()); what != nil {
			return what
		}
	}

	// If we haven't found our what and where yet check our thing's inventory
	if what = search(alias, t); what != nil {
		return what
	}

	// 404 - Not found :(
	return nil
}

func search(alias string, t has.Thing) (what has.Thing) {
	if a := attr.Inventory().Find(t); a != nil {
		if what = a.Search(alias); what != nil {
			return
		}
	}

	if a := attr.Narrative().Find(t); a != nil {
		if what = a.Search(alias); what != nil {
			return
		}
	}

	return nil
}

func CheckVetoes(cmd string, what has.Thing) has.Veto {
	if vetoes := attr.Vetoes().Find(what); vetoes != nil {
		if veto := vetoes.Check(cmd); veto != nil {
			return veto
		}
	}

	return nil
}
