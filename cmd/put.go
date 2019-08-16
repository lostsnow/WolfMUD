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

	if cWhat == nil {
		return
	}

	cInv := attr.FindInventory(cWhat)
	cName := attr.FindName(cWhat).Name("something")
	cCarried := cInv.Carried()

	who := attr.FindName(s.actor).TheName("someone")
	aInv := attr.FindInventory(s.actor)
	notifyObserver := false

	// Find items to put into container
	for _, match := range MatchAll(words, aInv.Contents()) {

		switch {
		case match.Unknown != "":
			s.msg.Actor.SendBad("You have no '", match.Unknown, "' to put into ", cName, ".")
			continue

		case match.NotEnough != "":
			s.msg.Actor.SendBad("You don't have that many '", match.NotEnough, "' to put into ", cName, ".")
			continue

		default:
			tWhat := match.Thing
			tName := attr.FindName(tWhat).Name("something")

			// Unless our name is Klein we can't put something inside itself! ;)
			if tWhat == cWhat {
				who := text.TitleFirst(who)
				s.msg.Actor.SendInfo("It might be interesting to put ", tName, " inside itself, but probably paradoxical as well.")
				s.msg.Observer.SendInfo(who, " seems to be trying to turn ", tName, " into a paradox.")
				continue
			}

			// Check put is not vetoed by item
			for _, vetoes := range attr.FindAllVetoes(tWhat) {
				if veto := vetoes.Check(s.actor, "PUT"); veto != nil {
					s.msg.Actor.SendBad(veto.Message())
					return
				}
			}

			// Remove item from actor and put it in the container
			aInv.Move(tWhat, cInv)

			// If item is not put into a carried container the item is now just
			// laying around so check if it should register for clean up
			if !cCarried {
				attr.FindCleanup(tWhat).Cleanup()
			}

			s.msg.Actor.SendGood("You put ", tName, " into ", cName, ".")
			notifyObserver = true
		}
	}

	if notifyObserver {
		s.msg.Observer.SendInfo("You see ", who, " put something into ", cName, ".")
	}

	s.ok = true
}

// findContainer looks in the actor's inventory then the location trying to
// find a matching valid container we can put items into. If a valid container
// cannot be found then container will be set to nil. Unprocessed words are
// returned for further matching. On failure appropriate message are sent to
// the actor and observers.
func (p put) findContainer(s *state) (container has.Thing, words []string) {

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
