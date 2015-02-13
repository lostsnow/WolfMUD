// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"strings"
)

// Syntax: #DUMP alias
func Dump(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "What do you want to dump?"
		return
	}

	var (
		name = aliases[0]

		what  has.Thing
		where has.Thing
	)

	// Work out where we are
	if a := attr.Locate().Find(t); a != nil {
		where = a.Where()
	}

	// Are we somewhere?
	if where != nil {
		// Search for item in inventory where we are
		if a := attr.Inventory().Find(where); a != nil {
			what = a.Search(name)
		}

		// If item not found in inventory try searching narratives
		if what == nil {
			if a := attr.Narrative().Find(where); a != nil {
				what = a.Search(name)
			}
		}
	}

	// If item still not found try our own inventory
	if what == nil {
		if a := attr.Inventory().Find(t); a != nil {
			what = a.Search(name)
		}
	}

	// If still not found try where we actually are
	if what == nil {
		if where != nil {
			if a := attr.Alias().Find(where); a != nil {
				if a.HasAlias(aliases[0]) {
					what = where
				}
			}
		}
	}

	if what == nil {
		msg = "Nothing with alias '" + aliases[0] + "' found to dump."
		return
	}

	msg = strings.Join(what.Dump(), "\n")
	return msg, true
}
