// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: TELL <who> <message> | TALK <who> <message>
func init() {
	addHandler(tell{}, "TELL")
	addHandler(tell{}, "TALK")
}

type tell cmd

func (tell) process(s *state) {
	if len(s.words) == 0 {
		switch s.cmd {
		case "TELL":
			s.msg.Actor.SendInfo("You go to tell someone something...")
		case "TALK":
			s.msg.Actor.SendInfo("Who did you want to talk to?")
		}
		return
	}

	for _, player := range s.where.Players() {
		if attr.FindAlias(player).HasAlias(s.words[0]) {
			s.participant = player
		}
	}

	if s.participant == nil {
		s.msg.Actor.SendBad("There is no '", s.words[0], "' here to talk to.")
		return
	}

	// Get all location inventories within 1 move of current location
	locations := attr.FindExits(s.where.Parent()).Within(1, s.where)

	// Try locking all of the locations we found
	lockAdded := false
	for _, d := range locations {
		for _, i := range d {
			if !s.CanLock(i) {
				s.AddLock(i)
				lockAdded = true
			}
		}
	}

	// If we added any locks return to the parser so we can relock
	if lockAdded {
		return
	}

	// Chop of stop words and alias from input
	for x, word := range s.input {
		if strings.ToUpper(word) == s.words[0] {
			s.input = s.input[x+1:]
			break
		}
	}

	name := attr.FindName(s.participant).TheName("someone")

	if len(s.input) == 0 {
		switch s.cmd {
		case "TELL":
			s.msg.Actor.SendBad("What did you want to tell ", name, "?")
		case "TALK":
			s.msg.Actor.SendInfo("What did you want say to ", name, "?")
		}
		return
	}

	who := text.TitleFirst(attr.FindName(s.actor).TheName("Someone"))
	msg := strings.Join(s.input, " ")

	s.msg.Actor.SendGood("You say to ", name, ": ", msg)
	s.msg.Participant.SendInfo(who, " says to you: ", msg)
	s.msg.Observer.SendInfo(who, " says to ", name, ": ", msg)

	// Notify observers in near by locations
	s.msg.Observers.Filter(locations[1]...).SendInfo("You hear talking nearby.")

	s.ok = true
	return
}
