// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: JUNK item
func init() {
	AddHandler(Junk, "JUNK")
}

func Junk(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to junk... something?")
		return
	}

	name := s.words[0]

	// Search for item we want to get in our inventory
	where := attr.FindInventory(s.actor)
	what := where.Search(name)

	// If not found check inventory where we are
	if what == nil {
		where = s.where
		what = where.Search(name)
	}

	// Still not found?
	if what == nil {
		s.msg.Actor.SendBad("You see no '", name, "' to junk.")
		return
	}

	// Get item's proper name
	name = attr.FindName(what).Name(name)

	// Is item a narrative?
	if attr.FindNarrative(what).Found() {
		s.msg.Actor.SendBad("You cannot junk ", name, ".")
		return
	}

	// Make sure we are locking the reset/respawn origin
	origin := attr.FindLocate(what).Origin()
	if !s.CanLock(origin) {
		s.AddLock(origin)
		return
	}

	// Check junking is not vetoed by the item
	if veto := attr.FindVetoes(what).Check("JUNK"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// Check if item is an Inventory. If it is check recusivly if its content can
	// be junked
	var check func(has.Thing) bool
	check = func(t has.Thing) bool {
		if i := attr.FindInventory(t); i.Found() {
			for _, t := range i.Contents() {
				if veto := attr.FindVetoes(t).Check("JUNK"); veto != nil {
					return false
				}
				if !check(t) {
					return false
				}
			}
		}
		return true
	}
	if !check(what) {
		s.msg.Actor.SendBad(text.TitleFirst(name), " seems to contain something that cannot be junked.")
		return
	}

	// Remove Thing from Inventory where it is. A respawn will be triggered if
	// Thing is spawnable and we will get the copy. Otherwise we get the original
	// back. If we get nil returned the Thing could not be removed for junking.
	if what = where.Remove(what); what == nil {
		s.msg.Actor.SendBad("For some reason you cannot junk ", name, ".")
		return
	}

	// Dispose of the Thing. If Thing was respawnable it will dispose of the
	// triggering copy.
	what.Dispose()

	who := attr.FindName(s.actor).Name("Someone")

	s.msg.Actor.SendGood("You junk ", name, ".")
	s.msg.Observer.SendInfo("You see ", who, " junk ", name, ".")
	s.ok = true
}
