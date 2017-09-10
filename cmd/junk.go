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

	// Search for item we want to junk in our inventory
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

	// Make sure we are locking the reset origin of the Thing to junk and the
	// origins of all of its content (recursively) if it has an Inventory.
	lc := len(s.locks)
	junkLockAll(s, what)
	if len(s.locks) != lc {
		return
	}

	// Check junking is not vetoed by the item
	if veto := attr.FindVetoes(what).Check("JUNK"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// Check if item is an Inventory. If it is check recusivly if its content can
	// be junked
	if !junkCheckVetoes(what) {
		s.msg.Actor.SendBad(text.TitleFirst(name), " seems to contain something that cannot be junked.")
		return
	}

	junkDispose(what)

	who := attr.FindName(s.actor).Name("Someone")

	s.msg.Actor.SendGood("You junk ", name, ".")
	s.msg.Observer.SendInfo("You see ", who, " junk ", name, ".")
	s.ok = true
}

// NOTE: junkLockAll is a temporary function until commands are made types and
// we can attach a lockAll method to the Junk type.
func junkLockAll(s *state, t has.Thing) {
	o := attr.FindLocate(t).Origin()
	s.AddLock(o)
	i := attr.FindInventory(t)
	for _, c := range i.Contents() {
		junkLockAll(s, c)
	}
}

// NOTE: junkCheckVetoes is a temporary function until commands are made types
// and we can attach a checkVetoes method to the Junk type.
func junkCheckVetoes(t has.Thing) bool {
	if i := attr.FindInventory(t); i.Found() {
		for _, t := range i.Contents() {
			if veto := attr.FindVetoes(t).Check("JUNK"); veto != nil {
				return false
			}
			if !junkCheckVetoes(t) {
				return false
			}
		}
	}
	return true
}

// NOTE: junkDispose is a temporary function until commands are made types and
// we can attach a dispose method to the Junk type.
func junkDispose(t has.Thing) {

	// Recurse into inventories
	i := attr.FindInventory(t)
	for _, c := range i.Contents() {
		junkDispose(c)
	}

	l := attr.FindLocate(t)
	if l.Origin() == nil {
		l.Where().Remove(t)
		t.Free()
		return
	}

	// When disposing of an item we can just call Reset on both resetable and
	// respawnable items. For respawnable items it avoids making a copy.
	if r := attr.FindReset(t); r.Found() {
		o := l.Origin()
		l.Where().Move(t, o)
		if l.Origin() != nil {
			o.Disable(t)
			r.Reset()
			return
		}
	}

	t.Free()
	return
}
