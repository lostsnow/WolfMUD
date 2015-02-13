// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Syntax: ( EXAMINE | EXAM ) item
func Examine(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "You examine this and that, find nothing special."
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

	// Was item to examine found?
	if what == nil {
		msg = "You see no '" + name + "' to examine."
		return
	}

	// Check examine is not vetoed by item
	if veto := CheckVetoes("EXAMINE", what); veto != nil {
		msg = veto.Message()
		return
	}

	buff := make([]byte, 0, 1024)

	if n := attr.Name().Find(what); n != nil {
		name = n.Name()
	}

	buff = append(buff, "You examine "...)
	buff = append(buff, name...)
	buff = append(buff, "."...)

	for _, d := range attr.Description().FindAll(what) {
		buff = append(buff, " "...)
		buff = append(buff, d.Description()...)
	}

	if i := attr.Inventory().Find(what); i != nil {
		buff = append(buff, " "...)
		buff = append(buff, i.Contents()...)
	}

	return string(buff), true
}
