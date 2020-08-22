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

// Syntax: WHISPER <who> <message>
func init() {
	addHandler(whisper{}, "WHISPER")
}

type whisper cmd

func (whisper) process(s *state) {
	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to whisper something to someone...")
		return
	}

	for _, player := range s.where.Players() {
		if attr.FindAlias(player).HasAlias(s.words[0]) {
			s.participant = player
		}
	}

	if s.participant == nil {
		s.msg.Actor.SendBad("There is no '", s.words[0], "' here to whisper to.")
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
		s.msg.Actor.SendBad("What did you want to whisper to ", name, "?")
		return
	}

	who := text.TitleFirst(attr.FindName(s.actor).TheName("Someone"))
	msg := strings.Join(s.input, " ")

	s.msg.Actor.SendGood("You whisper to ", name, ": ", msg)
	s.msg.Participant.SendInfo(who, " whispers to you: ", msg)
	s.msg.Observer.SendInfo(who, " whispers something to ", name, ".")

	s.ok = true
	return
}
