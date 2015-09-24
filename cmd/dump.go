// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"

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
		where has.Inventory
	)

	// Try our own inventory first for something matching the alias we are
	// looking for.
	if a := attr.FindInventory(t); a != nil {
		what = a.Search(name)
	}

	// If match not found work out where we are so we can search further
	if what == nil {
		if a := attr.FindLocate(t); a != nil {
			where = a.Where()
		}
	}

	// If match not found yet and we are not somewhere, we can't search any
	// further
	if what == nil && where == nil {
		msg = "You have nothing with alias '" +
			aliases[0] +
			"' to dump and nowhere else to search."

		return
	}

	// If match not found yet search where we are
	if what == nil {
		what = where.Search(name)
	}

	// If match not found try searching narratives
	location := where.Parent()
	if what == nil {
		if a := attr.FindNarrative(location); a != nil {
			what = a.Search(name)
		}
	}

	// If match still not found try the location itself - as opposed to it's
	// inventory and narratives.
	if what == nil {
		if a := attr.FindAlias(location); a != nil {
			if a.HasAlias(aliases[0]) {
				what = location
			}
		}
	}

	// If we havn't found a match by this stage we are not going to find one!
	if what == nil {
		msg = "Nothing with alias '" + aliases[0] + "' found to dump."
		return
	}

	msg = strings.Join(what.Dump(), "\n")
	return msg, true
}
