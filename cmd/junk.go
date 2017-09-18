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
	AddHandler(junk{}, "JUNK")
}

type junk cmd

func (j junk) process(s *state) {

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
	j.lockOrigins(s, what)
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
	if j.vetoed(what) {
		s.msg.Actor.SendBad(text.TitleFirst(name), " seems to contain something that cannot be junked.")
		return
	}

	j.dispose(what)

	who := attr.FindName(s.actor).Name("Someone")

	s.msg.Actor.SendGood("You junk ", name, ".")
	s.msg.Observer.SendInfo("You see ", who, " junk ", name, ".")
	s.ok = true
}

// lockOrigins adds locks for the origin of the passed Thing and the origins of
// all of its Inventory content recursively.
func (j junk) lockOrigins(s *state, t has.Thing) {
	o := attr.FindLocate(t).Origin()
	s.AddLock(o)
	i := attr.FindInventory(t)
	for _, c := range i.Contents() {
		j.lockOrigins(s, c)
	}
}

// vetoed checks the Inventory content (recursively) of the passed Thing to see
// if any of the content vetoes the JUNK command. If anything vetoes the JUNK
// command returns true otherwise returns false.
func (j junk) vetoed(t has.Thing) bool {
	if i := attr.FindInventory(t); i.Found() {
		for _, t := range i.Contents() {
			if veto := attr.FindVetoes(t).Check("JUNK"); veto != nil {
				return true
			}
			if j.vetoed(t) {
				return true
			}
		}
	}
	return false
}

// dispose takes a thing out of play. If the Thing has a Reset attribute and an
// origin it will be schedued for a reset. Otherwise the Thing will be removed
// and released for garbase collection.
func (j junk) dispose(t has.Thing) {

	// Recurse into inventories and junk content
	i := attr.FindInventory(t)
	for _, c := range i.Contents() {
		j.dispose(c)
	}

	l := attr.FindLocate(t)
	o := l.Origin()
	r := attr.FindReset(t)

	// If Thing has no reset or origin remove it and free for garbage collection
	if !r.Found() || o == nil {
		l.Where().Remove(t)
		t.Free()
		return
	}

	// Move Thing to its origin ready for reset. If Thing is spawnable and we get
	// a copy of the Thing we have to dispose of the copy. In which case the
	// original will already have been disabled and registered for a reset.
	c := l.Where().Move(t, o)
	if c.UID() != t.UID() {
		j.dispose(c)
		return
	}

	o.Disable(t)
	r.Reset()

	return
}
