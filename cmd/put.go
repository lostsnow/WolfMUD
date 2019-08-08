// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: PUT item... container
func init() {
	addHandler(put{}, "PUT")
}

type put cmd

func (put) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to put something into something else...")
		return
	}

	// Look for container to put things into
	aInv := attr.FindInventory(s.actor)
	matches, words := Match(s.words, aInv.Contents(), s.where.Everything())

	// If multiple containers found which one do we want?
	if len(matches) > 1 {
		s.msg.Actor.SendBad("Which container did you mean?")
		for _, match := range matches {
			s.msg.Actor.Send("  ", attr.FindName(match).Name("something"))
		}
		return
	}

	// Was a single container actually found?
	switch match := matches[0]; {
	case match.Unknown != "":
		if len(words) == 0 {
			s.msg.Actor.SendBad("You have no '", match.Unknown, "' to put into anything.")
		} else {
			s.msg.Actor.SendBad("You see no '", match.Unknown, "' to put things into.")
		}
		return

	case match.NotEnough != "":
		if len(words) == 0 {
			s.msg.Actor.SendBad("You don't have that many '", match.NotEnough, "' to put into anything.")
		} else {
			s.msg.Actor.SendBad("You don't see that many '", match.NotEnough, "' to put things into.")
		}
		return
	}

	cWhat := matches[0].Thing
	cName := attr.FindName(cWhat).Name("something")
	cInv := attr.FindInventory(cWhat)
	cCarried := cInv.Carried()

	// If nothing else specified assume this is an item and we have no container
	if len(words) == 0 {
		s.msg.Actor.SendBad("What did you want to put ", cName, " into?")
		return
	}

	// Is the container actually a container and something we can put things into?
	if !cInv.Found() {
		name := text.TitleFirst(cName)
		s.msg.Actor.SendBad(name, " isn't something you can put things in.")
		return
	}

	// Check putting things into the container not vetoed by container
	for _, vetoes := range attr.FindAllVetoes(cWhat) {
		if veto := vetoes.Check(s.actor, "PUTIN"); veto != nil {
			s.msg.Actor.SendBad(veto.Message())
			return
		}
	}

	who := attr.FindName(s.actor).TheName("someone")
	notifyObserver := false

	// Find items to put into container
	for _, match := range MatchAll(words, aInv.Contents()) {

		switch {
		case match.Unknown != "":
			s.msg.Actor.SendBad("You have no '", match.Unknown, "' to put into anything.")
			return

		case match.NotEnough != "":
			s.msg.Actor.SendBad("You don't have that many '", match.NotEnough, "' to put into ", cName, ".")
			return

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
