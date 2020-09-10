// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: $ACT <description>
func init() {
	addHandler(prompt{}, "/prompt")
}

type prompt cmd

func (prompt) process(s *state) {

	if len(s.words) == 0 {
		old := attr.FindPlayer(s.actor).SetPromptStyle(has.StyleNone)
		switch old {
		case has.StyleNone:
			s.msg.Actor.SendInfo("Prompt is currently set to none.")
		case has.StyleBrief:
			s.msg.Actor.SendInfo("Prompt is currently set to brief.")
		case has.StyleShort:
			s.msg.Actor.SendInfo("Prompt is currently set to short.")
		case has.StyleLong:
			s.msg.Actor.SendInfo("Prompt is currently set to long.")
		default:
			s.msg.Actor.SendInfo("Prompt is currently unknown.")
		}
		attr.FindPlayer(s.actor).SetPromptStyle(old)
		s.ok = true
		return
	}

	old := attr.FindPlayer(s.actor).SetPromptStyle(has.StyleNone)

	switch s.words[0] {
	case "NONE":
		s.msg.Actor.SendInfo("Prompt set to none.")
		old = has.StyleNone
	case "BRIEF":
		s.msg.Actor.SendInfo("Prompt set to brief.")
		old = has.StyleBrief
	case "SHORT":
		s.msg.Actor.SendInfo("Prompt set to short.")
		old = has.StyleShort
	case "LONG":
		s.msg.Actor.SendInfo("Prompt set to long.")
		old = has.StyleLong
	default:
		s.msg.Actor.SendInfo("Invalid prompt given.")
	}

	attr.FindPlayer(s.actor).SetPromptStyle(old)

	s.ok = true
}
