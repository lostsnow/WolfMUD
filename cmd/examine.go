// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: ( EXAMINE | EXAM ) item
func init() {
	addHandler(examine{}, "EXAM", "EXAMINE")
}

type examine cmd

func (examine) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You examine this and that, find nothing special.")
		return
	}

	// Find matching item at location or held by actor
	matches, words := Match(
		s.words,
		s.where.Everything(),
		attr.FindInventory(s.actor).Contents(),
	)
	match := matches[0]
	mark := s.msg.Actor.Len()

	switch {
	case len(words) != 0: // Not exact match?
		name := strings.Join(s.words, " ")
		s.msg.Actor.SendBad("You see no '", name, "' to examine.")

	case len(matches) != 1: // More than one match?
		s.msg.Actor.SendBad("You can only examine one thing at a time.")

	case match.Unknown != "":
		s.msg.Actor.SendBad("You see no '", match.Unknown, "' to examine.")

	case match.NotEnough != "":
		s.msg.Actor.SendBad("There are not that many '", match.NotEnough, "' to examine.")

	}

	// If we sent an error to the actor return now
	if mark != s.msg.Actor.Len() {
		return
	}

	what := match.Thing

	// Check examine is not vetoed by item or location
	for _, t := range []has.Thing{what, s.where.Parent()} {
		for _, vetoes := range attr.FindAllVetoes(t) {
			if veto := vetoes.Check(s.actor, "EXAMINE", "EXAM"); veto != nil {
				s.msg.Actor.SendBad(veto.Message())
				return
			}
		}
	}

	name := attr.FindName(what).TheName("something") // Get item's proper name

	s.msg.Actor.SendGood("You examine ", name, ".", text.Reset, "\n")

	for _, a := range attr.FindAllDescription(what) {
		s.msg.Actor.Append(a.Description())
	}

	isPlayer := attr.FindPlayer(what).Found()

	// If examining a player they become the participant
	if isPlayer {
		s.participant = what
	}

	// BUG(diddymus): If you examine another player you can see their inventory
	// items. For now we only describe the inventory if not examining a player.
	if !isPlayer {
		if l := attr.FindInventory(what).List(); l != "" {
			s.msg.Actor.Append(l)
		}
	}

	who := attr.FindName(s.actor).TheName("Someone")
	who = text.TitleFirst(who)
	name = attr.FindName(what).Name(name)

	s.msg.Participant.SendInfo(who, " studies you.")

	if !attr.FindLocate(what).Where().Carried() {
		s.msg.Observer.SendInfo(who, " studies ", name, ".")
	} else {
		s.msg.Observer.SendInfo(who, " studies ", name, " they are carrying.")
	}

	s.ok = true
}
