// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"strings"
	"sync"

	"code.wolfmud.org/WolfMUD.git/mailbox"
)

// World contains all of the locations for the current game world. It is
// protected by the BWL (Big World Lock).
var BWL sync.Mutex
var World map[string]*Thing

// WorldStart contains a list of references to starting locations
var WorldStart []string

type state struct {
	actor *Thing
	cmd   string
	word  []string
	buff  *strings.Builder
}

var newline = []byte("\n")

func NewState(t *Thing) *state {
	return &state{actor: t, buff: &strings.Builder{}}
}

func (s *state) Parse(input string) (cmd string) {
	if input = strings.TrimSpace(input); len(input) != 0 {
		s.parse(input)
	}
	mailbox.Send(s.actor.As[UID], s.buff.String())
	s.buff.Reset()
	return s.cmd
}

func (s *state) parse(input string) {
	s.word = strings.Fields(strings.ToUpper(input))
	s.cmd, s.word = s.word[0], s.word[1:]

	if command, ok := commands[s.cmd]; ok {
		// Stop the world for everyone else...
		BWL.Lock()
		defer BWL.Unlock()
		command(s)
	} else {
		s.Msg("Eh?")
	}
}

// Msg sends a message to the actor. If anything has already been sent to the
// actor for this command a line-feed is automatically added at the beginning
// of the message.
func (s *state) Msg(text ...string) {
	if s.buff.Len() > 0 {
		s.buff.Write(newline)
	}
	s.MsgAppend(text...)
}

// MsgAppend sends a message to the actor. Unlike Msg, MsgAppend never adds
// line-feeds automatically - which can be useful when building up messages in
// stages.
func (s *state) MsgAppend(text ...string) {
	for _, t := range text {
		s.buff.WriteString(t)
	}
}
