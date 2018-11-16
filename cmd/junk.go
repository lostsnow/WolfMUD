// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/event"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: JUNK item
func init() {
	addHandler(junk{}, "JUNK")
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
	if veto := attr.FindVetoes(what).Check(s.actor, "JUNK"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// Check if item is an Inventory. If it is check recusivly if its content can
	// be junked
	if j.vetoed(s.actor, what) {
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
// all of the content of its Inventory - recursively.
func (j junk) lockOrigins(s *state, t has.Thing) {
	s.AddLock(attr.FindLocate(t).Origin())
	for _, c := range attr.FindInventory(t).Contents() {
		j.lockOrigins(s, c)
	}
}

// vetoed checks the content of an Inventory (recursively) of the passed Thing
// to see if any of the content vetoes the JUNK command. If anything vetoes the
// JUNK command returns true otherwise returns false.
func (j junk) vetoed(actor has.Thing, t has.Thing) bool {
	for _, t := range attr.FindInventory(t).Contents() {
		if attr.FindVetoes(t).Check(actor, "JUNK") != nil {
			return true
		}
		return j.vetoed(actor, t)
	}
	return false
}

// dispose takes a thing out of play. If the Thing is collectable it will be
// removed and released for garbage collection. If the Thing is not collectable
// a reset will be scheduled.
func (j junk) dispose(t has.Thing) {

	// Recurse into inventories and dispose of the content
	for _, c := range attr.FindInventory(t).Contents() {
		j.dispose(c)
	}

	l := attr.FindLocate(t)
	w := l.Where()
	o := l.Origin()
	r := attr.FindReset(t)

	attr.FindAction(t).Abort()
	attr.FindCleanup(t).Abort()

	// If Thing is collectable remove it and free for garbage collection
	if t.Collectable() {
		w.Disable(t)
		w.Remove(t)
		t.Free()
		return
	}

	// Move Thing to its origin and disable it, as it is out of play
	w.Move(t, o)
	o.Disable(t)

	// If we don't have a reset attribute then invoke a "$RESET" on the fly to
	// force a reset. The reset will happen almost immediately and players will
	// see any relevant reset messages.
	if !r.Found() {
		event.Queue(t, "$RESET", 0, 0)
		return
	}

	// Register for a reset use reset attribute
	r.Reset()

	return
}
