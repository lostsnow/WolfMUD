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
func init() {
	AddHandler(Dump, "#DUMP")
}

func Dump(s *state) {

	if len(s.words) == 0 {
		s.msg.actor.WriteString("What do you want to dump?")
		return
	}

	var (
		name = s.words[0]

		what     has.Thing
		location has.Thing
	)

	// If we can, search where we are
	if s.where != nil {
		what = s.where.Search(name)
		location = s.where.Parent()
	}

	// If item still not found see if we can search narratives
	if what == nil && location != nil {
		what = attr.FindNarrative(location).Search(name)
	}

	// If item still not found try our own inventory
	if what == nil {
		what = attr.FindInventory(s.actor).Search(name)
	}

	// If match still not found try the location itself - as opposed to it's
	// inventory and narratives.
	if what == nil && location != nil {
		if attr.FindAlias(location).HasAlias(s.words[0]) {
			what = location
		}
	}

	// If item still not found try the actor - normally we would find the actor
	// in the location's inventory, assuming the actor is somewhere. If the actor
	// is nowhere we have to check it specifically.
	if what == nil && location == nil {
		if attr.FindAlias(s.actor).HasAlias(s.words[0]) {
			what = s.actor
		}
	}

	// Was item to dump eventually found?
	if what == nil {
		s.msg.actor.WriteJoin("There is nothing with alias '", s.words[0], "' to dump.")
		return
	}

	s.msg.actor.WriteString(strings.Join(what.Dump(), "\n"))
	s.ok = true
}
