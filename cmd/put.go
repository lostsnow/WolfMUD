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

// Syntax: PUT item... container
func init() {
	addHandler(put{}, "PUT")
}

type put cmd

func (p put) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to put something into something else...")
		return
	}

	cWhat, words := p.findContainer(s)

	// If no container we can't continue
	if cWhat == nil {
		return
	}

	aInv := attr.FindInventory(s.actor)
	cInv := attr.FindInventory(cWhat)
	cName := attr.FindName(cWhat).TheName("something")

	notifyObserver := false

	// Match items to put into container
	for _, match := range MatchAll(words, aInv.Contents()) {

		tWhat := p.findItem(s, cWhat, match)

		// If item not matched move onto next item
		if tWhat == nil {
			continue
		}

		// Move the item from the actor's inventory to the container
		aInv.Move(tWhat, cInv)

		// If item is not put into a carried container the item is now just
		// laying around so check if it should register for clean up
		if !cInv.Carried() {
			attr.FindCleanup(tWhat).Cleanup()
		}

		tName := attr.FindName(tWhat).TheName("something")
		s.msg.Actor.SendGood("You put ", tName, " into ", cName, ".")
		notifyObserver = true
	}

	if notifyObserver {
		who := attr.FindName(s.actor).TheName("someone")
		cName := attr.FindName(cWhat).Name("something")
		s.msg.Observer.SendInfo("You see ", who, " put something into ", cName, ".")
	}

	s.ok = true
}

// findContainer looks in the actor's inventory then the location trying to
// find a matching valid container we can put items into. If a valid container
// cannot be found then container will be set to nil. Unprocessed words are
// returned for further matching. On failure appropriate message are sent to
// the actor and observers.
func (put) findContainer(s *state) (container has.Thing, words []string) {

	matches, words := Match(
		s.words,
		attr.FindInventory(s.actor).Contents(),
		s.where.Everything(),
	)
	what := matches[0]
	noItems := len(words) == 0
	mark := s.msg.Actor.Len()

	switch {
	// If we only have "PUT item" and item unknown
	case noItems && what.Unknown != "":
		s.msg.Actor.SendBad("You have no '", what.Unknown, "' to put into anything.")

	// If we have "PUT items... container" and container is unknown
	case what.Unknown != "":
		s.msg.Actor.SendBad("You see no '", what.Unknown, "' to put things into.")

	// If we only have "PUT item" and not enough of item
	case noItems && what.NotEnough != "":
		s.msg.Actor.SendBad(
			"You don't have that many '", what.NotEnough, "' to put into anything.",
		)

	// If we have "PUT items... container" and not enough of container
	case what.NotEnough != "":
		s.msg.Actor.SendBad(
			"You don't see that many '", what.NotEnough, "' to put things into.",
		)

	// If we have "PUT item..." and more than one match assume no container
	case noItems && len(matches) > 1:
		s.msg.Actor.SendBad("You go to put things into... something?")

	// If we have "PUT item... container" and more than one container match
	case len(matches) > 1:
		s.msg.Actor.SendBad("You can only put things into one container at a time.")
	}

	// If we sent an error to the actor return now
	if mark != s.msg.Actor.Len() {
		return nil, words
	}

	// Something has been matched so try to get its name and inventory
	name := attr.FindName(what).Name("something")
	inv := attr.FindInventory(what)

	switch {
	// If we have "PUT item" and match is not actually a container
	case noItems && !inv.Found():
		s.msg.Actor.SendBad("What did you want to put ", name, " into?")

	// If we have "PUT item" and match actually is a container
	case noItems && inv.Found():
		s.msg.Actor.SendBad("Did you want to put something into ", name, "?")

	// Is the container actually a container and something we can put things into?
	case !inv.Found():
		s.msg.Actor.SendBad(
			text.TitleFirst(name), " isn't something you can put things in.",
		)
	}

	// If we sent an error to the actor return now
	if mark != s.msg.Actor.Len() {
		return nil, words
	}

	// Check putting things into the container not vetoed by container
	for _, vetoes := range attr.FindAllVetoes(what) {
		if veto := vetoes.Check(s.actor, "PUTIN"); veto != nil {
			s.msg.Actor.SendBad(veto.Message())
			return nil, words
		}
	}

	return what.Thing, words
}

// findItem checks the match passed to it to see if it contains an item that
// can be placed into the specified container. Returns the item from the match
// if it is valid else nil. On failure appropriate message are sent to the
// actor and observers.
func (put) findItem(s *state, container has.Thing, match Result) has.Thing {

	if match.Unknown != "" {
		name := attr.FindName(container).TheName("something")
		s.msg.Actor.SendBad(
			"You have no '", match.Unknown, "' to put into ", name, ".",
		)
		return nil
	}

	if match.NotEnough != "" {
		name := attr.FindName(container).TheName("something")
		s.msg.Actor.SendBad(
			"You don't have that many '", match.NotEnough, "' to put into ", name, ".",
		)
		return nil
	}

	what := match.Thing

	// Unless our name is Klein we can't put something inside itself! ;)
	if what == container {
		who := text.TitleFirst(attr.FindName(s.actor).TheName("someone"))
		name := attr.FindName(what).Name("something")

		s.msg.Actor.SendInfo(
			"It might be interesting to put ", name,
			" inside itself, but probably paradoxical as well.",
		)

		s.msg.Observer.SendInfo(
			who, " seems to be trying to turn ", name, " into a paradox.",
		)

		return nil
	}

	// Check put is not vetoed by item
	for _, vetoes := range attr.FindAllVetoes(what) {
		if veto := vetoes.Check(s.actor, "PUT"); veto != nil {
			s.msg.Actor.SendBad(veto.Message())
			return nil
		}
	}

	return what
}
