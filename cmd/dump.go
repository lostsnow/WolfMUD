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
		s.msg.Actor.Send("What do you want to dump?")
		return
	}

	name := s.words[0]

	var what has.Thing

	// If we can, search where we are
	if s.where != nil {
		what = s.where.Search(name)
	}

	// If item still not found try our own inventory
	if what == nil {
		what = attr.FindInventory(s.actor).Search(name)
	}

	// If match still not found try the location itself - as opposed to it's
	// inventory and narratives.
	if what == nil && s.where != nil {
		if attr.FindAlias(s.where.Parent()).HasAlias(name) {
			what = s.where.Parent()
		}
	}

	// If item still not found  and we are nowhere try the actor - normally we
	// would find the actor in the location's inventory, assuming the actor is
	// somewhere. If the actor is nowhere we have to check it specifically.
	if what == nil && s.where == nil {
		if attr.FindAlias(s.actor).HasAlias(name) {
			what = s.actor
		}
	}

	// Was item to dump eventually found?
	if what == nil {
		s.msg.Actor.Send("There is nothing with alias '", name, "' to dump.")
		return
	}

	s.msg.Actor.Send(strings.Join(what.Dump(), "\n"))
	s.ok = true
}
