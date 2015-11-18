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
func Dump(s *state) {

	if len(s.words) == 0 {
		s.msg.actor.WriteString("What do you want to dump?")
		return
	}

	var (
		name = s.words[0]

		what  has.Thing
		where has.Inventory
	)

	// Try our own inventory first for something matching the alias we are
	// looking for.
	if a := attr.FindInventory(s.actor); a != nil {
		what = a.Search(name)
	}

	// If match not found work out where we are so we can search further
	if what == nil {
		if a := attr.FindLocate(s.actor); a != nil {
			where = a.Where()
		}
	}

	// If match not found yet and we are not somewhere, we can't search any
	// further
	if what == nil && where == nil {
		s.msg.actor.WriteString("You have nothing with alias '" + s.words[0] + "' to dump and nowhere else to search.")

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
			if a.HasAlias(s.words[0]) {
				what = location
			}
		}
	}

	// If we havn't found a match by this stage we are not going to find one!
	if what == nil {
		s.msg.actor.WriteString("Nothing with alias '" + s.words[0] + "' found to dump.")
		return
	}

	s.msg.actor.WriteString(strings.Join(what.Dump(), "\n"))
	s.ok = true
}
