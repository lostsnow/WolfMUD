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
var World Things

// WorldStart contains a list of references to starting locations
var WorldStart []string

type state struct {
	actor *Thing
	buf   map[string]*strings.Builder
	cmd   string
	input string
	word  []string
}

var newline = []byte("\n")

func NewState(t *Thing) *state {
	return &state{actor: t, buf: make(map[string]*strings.Builder)}
}

func (s *state) Parse(input string) (cmd string) {
	if input = strings.TrimSpace(input); len(input) != 0 {
		s.parse(input)
	}
	return s.cmd
}

func (s *state) parse(input string) {
	s.word = strings.Fields(strings.ToUpper(input))
	s.cmd, s.word = s.word[0], s.word[1:]
	s.input = strings.TrimSpace(input[len(s.cmd):])

	// Stop the world for everyone else...
	BWL.Lock()
	defer BWL.Unlock()

	if command, ok := commands[s.cmd]; ok {
		savedDA := s.actor.As[DynamicAlias]
		s.actor.As[DynamicAlias] = "SELF"
		command(s)
		s.actor.As[DynamicAlias] = savedDA
	} else {
		s.Msg(s.actor, "Eh?")
	}

	s.mailman()
}

// mailman delivers queued messages to player's mailboxes. Messages can be
// queued for a specific player or for a location. If queued for a location,
// messages will be sent to all player at the location - unless they have
// received a specific message.
//
// Note that even though commands are processed under the BRL mailboxes can be
// deleted at anytime due to network errors. This is not a problem, if the UID
// for a buffer is not for an existing mailbox or location it will be ignored
// and cleaned up.
func (s *state) mailman() {

	for uid, buf := range s.buf {
		// Send to specific players - race between Exists & Send is okay
		if mailbox.Exists(uid) {
			mailbox.Send(uid, buf.String())
			continue
		}
		// Send to players at location, omitting players that are receiving
		// specific messages.
		if where := World[uid]; where != nil {
			for uid := range where.Who {
				if s.buf[uid] == nil {
					mailbox.Send(uid, buf.String())
				}
			}
		}
	}

	// Cleanup buffers
	for uid, buf := range s.buf {
		buf.Reset()
		delete(s.buf, uid)
	}
}

// Msg queues a message for the specified receiver. The receiver may be a
// player or location. If a player is specified the message is only sent to
// that player. If a loction is specified then the message is sent to all
// players at that location that have not received a specific message. All
// messages are sent once the current player commands completes. Msg may be
// called multiple times for the same recipient for a command, in which case
// the messages will be sent as a single delivery. Msg will always start the
// given text on a new line. To append text to the end of a message, without
// starting on a new line, use MsgAppend.
func (s *state) Msg(recipient *Thing, text ...string) {
	uid := recipient.As[UID]
	if s.buf[uid] == nil {
		s.buf[uid] = &strings.Builder{}
		if uid != s.actor.As[UID] {
			s.buf[uid].Write(newline)
		}
	} else {
		s.buf[uid].Write(newline)
	}
	for _, t := range text {
		s.buf[uid].WriteString(t)
	}
}

// MsgAppend works the same as Msg, but does not force a line-feed to be added
// before appending the text. This can be used to build messages a piece at a
// time. It is safe to call MsgAppend for a recipient, even if Msg has not been
// called first.
func (s *state) MsgAppend(recipient *Thing, text ...string) {
	uid := recipient.As[UID]
	if s.buf[uid] == nil {
		s.Msg(recipient, text...)
		return
	}
	for _, t := range text {
		s.buf[uid].WriteString(t)
	}
}
