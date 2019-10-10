// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: TAKE item... container
func init() {
	addHandler(take{}, "TAKE")
}

type take struct {
	cmd
	rummage bool // Has rummage message been seen already?
	trouble bool // Has trouble message been seen already?
}

func (t take) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to take something out of something else...")
		return
	}

	cWhat, words := t.findContainer(s)

	// If no container we can't continue
	if cWhat == nil {
		return
	}

	aInv := attr.FindInventory(s.actor)
	cInv := attr.FindInventory(cWhat)
	cName := attr.FindName(cWhat).TheName("something")

	notifyObserver := false

	// Match items to take from container
	for _, match := range MatchAll(words, cInv.Everything()) {

		tWhat := t.findItem(s, cWhat, match)

		// If item not matched move onto next item
		if tWhat == nil {
			continue
		}

		// Cancel any Cleanup or Action events
		attr.FindCleanup(tWhat).Abort()
		attr.FindAction(tWhat).Abort()

		// Check if item respawns when taken, if it does use spawned copy
		if s := attr.FindReset(tWhat).Spawn(); s != nil {
			tWhat = s
		}

		// Move the item from container to the actor's inventory
		cInv.Move(tWhat, aInv)

		tName := attr.FindName(tWhat).TheName("something")
		s.msg.Actor.SendGood("You take ", tName, " out of ", cName, ".")
		notifyObserver = true
	}

	if notifyObserver {
		who := attr.FindName(s.actor).TheName("someone")
		cName := attr.FindName(cWhat).Name("something")
		s.msg.Observer.SendInfo("You see ", who, " take something out of ", cName, ".")
	}

	s.ok = true
}

// findContainer looks in the actor's inventory then the location trying to
// find a matching valid container we can take items from. If a valid container
// cannot be found then container will be set to nil. Unprocessed words are
// returned for further matching. On failure appropriate message are sent to
// the actor and observers.
func (take) findContainer(s *state) (container has.Thing, words []string) {

	matches, words := Match(
		s.words,
		attr.FindInventory(s.actor).Contents(),
		s.where.Everything(),
	)
	what := matches[0]
	noItems := len(words) == 0
	mark := s.msg.Actor.Len()

	switch {
	// If we only have "TAKE item" and item unknown
	case noItems && what.Unknown != "":
		s.msg.Actor.SendBad("What did you want to take '", what.Unknown, "' out of?")

	// If we have "TAKE items... container" and container is unknown
	case what.Unknown != "":
		s.msg.Actor.SendBad("You see no '", what.Unknown, "' to take things out of.")

	// If we only have "TAKE item" and not enough of item
	case noItems && what.NotEnough != "":
		s.msg.Actor.SendBad(
			"What did you want to take '", what.NotEnough, "' out of?",
		)

	// If we have "TAKE items... container" and not enough of container
	case what.NotEnough != "":
		s.msg.Actor.SendBad(
			"You don't see that many '", what.NotEnough, "' to take things out of.",
		)

	// If we have "TAKE item..." and more than one match assume no container
	case noItems && len(matches) > 1:
		s.msg.Actor.SendBad("You go to take things out of... something?")

	// If we have "TAKE item... container" and more than one container match
	case len(matches) > 1:
		s.msg.Actor.SendBad("You can only take things from one container at a time.")
	}

	// If we sent an error to the actor return now
	if mark != s.msg.Actor.Len() {
		return nil, words
	}

	// A container has been matched so try to get its name and inventory
	name := attr.FindName(what).TheName("something")
	inv := attr.FindInventory(what)

	switch {
	// If we have "TAKE item" and match is not actually a container
	case noItems && !inv.Found():
		s.msg.Actor.SendBad("What did you want to take ", name, " from?")

	// If we have "TAKE item" and match actually is a container
	case noItems && inv.Found():
		s.msg.Actor.SendBad("Did you want to take something from ", name, "?")

	// Is the container actually a container and something we can take things
	// from?
	case !inv.Found():
		s.msg.Actor.SendBad("You cannot take anything from ", name, ".")
	}

	// If we sent an error to the actor return now
	if mark != s.msg.Actor.Len() {
		return nil, words
	}

	// Check taking things from the container not vetoed by container
	for _, vetoes := range attr.FindAllVetoes(what) {
		if veto := vetoes.Check(s.actor, "TAKEOUT"); veto != nil {
			s.msg.Actor.SendBad(veto.Message())
			return nil, words
		}
	}

	return what.Thing, words
}

// findItem checks the match passed to it to see if it contains an item that
// can be placed into the specified container. Returns the item from the match
// if it is valid else nil. On failure appropriate messages are sent to the
// actor and observers.
func (t *take) findItem(s *state, container has.Thing, match Result) has.Thing {

	if match.Unknown != "" {
		cName := text.TitleFirst(attr.FindName(container).TheName("something"))
		s.msg.Actor.SendBad(
			cName, " does not seem to contain '", match.Unknown, "'.",
		)

		if !t.rummage {
			who := attr.FindName(s.actor).TheName("someone")
			cName = attr.FindName(container).Name("something")
			s.msg.Observer.SendInfo("You see ", who, " rummage around in ", cName, ".")
			t.rummage = true
		}

		return nil
	}

	if match.NotEnough != "" {
		cName := attr.FindName(container).TheName("something")
		s.msg.Actor.SendBad(
			"There are not that many '", match.NotEnough, "' to take from ", cName, ".",
		)
		return nil
	}

	what := match.Thing
	where := attr.FindInventory(s.actor)

	// Check that the thing doing the taking can carry the item.
	//
	// NOTE: We could just drop the item on the floor if it can't be carried.
	if !where.Found() {
		cName := attr.FindName(container).TheName("something")
		tName := attr.FindName(what).TheName("something")
		s.msg.Actor.SendBad("You have nowhere to put ", tName, " if you remove it from ", cName, ".")
		return nil
	}

	// Check take is not vetoed by item
	for _, vetoes := range attr.FindAllVetoes(what) {
		if veto := vetoes.Check(s.actor, "TAKE"); veto != nil {
			s.msg.Actor.SendBad(veto.Message())
			return nil
		}
	}

	// If item is a narrative we can't take it. We do this check after the veto
	// checks as the vetos could give us a better message/reson for not being
	// able to take the item.
	if attr.FindNarrative(what).Found() {
		cName := attr.FindName(container).Name("something")
		cTheName := attr.FindName(container).TheName("something")
		tName := attr.FindName(what).TheName("something")
		s.msg.Actor.SendBad("For some reason you cannot take ", tName, " from ", cTheName, ".")

		if !t.trouble {
			who := attr.FindName(s.actor).TheName("someone")
			s.msg.Observer.SendInfo("You see ", who, " having trouble removing something from ", cName, ".")
			t.trouble = true
		}
		return nil
	}

	return what
}
