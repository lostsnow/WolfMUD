// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"

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

	// Find matching item at location or held by actor
	//
	// NOTE: s.where may be nil as in the case of a cleanup where the actor is
	// the location doing the junking.
	invs := [][]has.Thing{
		attr.FindInventory(s.actor).Contents(),
	}
	if s.where != nil {
		invs = append(invs, s.where.Everything())
	}

	matches, words := Match(s.words, invs...)
	match := matches[0]
	mark := s.msg.Actor.Len()

	switch {
	case len(words) != 0: // Not exact match?
		name := strings.Join(s.words, " ")
		s.msg.Actor.SendBad("You see no '", name, "' to junk.")

	case len(matches) != 1: // More than one match?
		s.msg.Actor.SendBad("You can only junk one thing at a time.")

	case match.Unknown != "":
		s.msg.Actor.SendBad("You see no '", match.Unknown, "' to junk.")

	case match.NotEnough != "":
		s.msg.Actor.SendBad("There are not that many '", match.NotEnough, "' to junk.")

	}

	// If we sent an error to the actor return now
	if mark != s.msg.Actor.Len() {
		return
	}

	what := match.Thing
	name := attr.FindName(what).TheName("something") // Get item's proper name

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
	for _, vetoes := range attr.FindAllVetoes(what) {
		if veto := vetoes.Check(s.actor, s.cmd); veto != nil {
			s.msg.Actor.SendBad(veto.Message())
			return
		}
	}

	// Check if item is an Inventory. If it is check recusivly if its content can
	// be junked
	if j.vetoed(s.actor, what) {
		s.msg.Actor.SendBad(text.TitleFirst(name), " seems to contain something that cannot be junked.")
		return
	}

	s.msg.Actor.SendGood("You junk ", name, ".")
	name = attr.FindName(what).Name("something")

	j.dispose(what)

	who := attr.FindName(s.actor).TheName("someone")
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
	for _, c := range attr.FindInventory(t).Disabled() {
		j.lockOrigins(s, c)
	}
}

// vetoed checks the content of an Inventory (recursively) of the passed Thing
// to see if any of the content vetoes the JUNK command. If anything vetoes the
// JUNK command returns true otherwise returns false.
func (j junk) vetoed(actor has.Thing, t has.Thing) bool {
	for _, t := range attr.FindInventory(t).Contents() {
		for _, veto := range attr.FindAllVetoes(t) {
			if veto.Check(actor, "JUNK") != nil {
				return true
			}
		}
		if j.vetoed(actor, t) {
			return true
		}
	}
	return false
}

// dispose takes a thing out of play. If the Thing is collectable it will be
// removed and released for garbage collection. If the Thing is not collectable
// a reset will be scheduled.
func (j junk) dispose(t has.Thing) {

	attr.FindAction(t).Abort()
	attr.FindCleanup(t).Abort()
	attr.FindReset(t).Abort()

	// Recurse into inventories and dispose of the content
	for _, c := range attr.FindInventory(t).Contents() {
		j.dispose(c)
	}
	for _, c := range attr.FindInventory(t).Disabled() {
		j.dispose(c)
	}

	l := attr.FindLocate(t)
	w := l.Where()
	o := l.Origin()
	r := attr.FindReset(t)

	// If Thing is spawned (it's a copy) or has no origin then remove it and free
	// for garbage collection
	if r.IsSpawned() || o == nil || !o.Found() {
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
		event.Queue(o.Parent(), "$RESET "+t.UID(), 0, 0)
		return
	}

	// Register for a reset use reset attribute
	r.Reset()

	return
}
