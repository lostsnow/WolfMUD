// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"strings"
	"sync"

	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/mailbox"
)

// World contains all of the top level locations for the current game world.
// WorldStart only contains valid player starting locations. Both are protected
// by the BWL (Big World Lock).
var (
	BWL        sync.Mutex
	World      Things   // All top level locations
	WorldStart []*Thing // Starting locations
)

type pkgConfig struct {
	crowdSize int // Represents minimum number of players considered a crowd
}

// cfg setup by Config and should be treated as immutable and not changed.
var cfg pkgConfig

// Config sets up package configuration for settings that can't be constants.
// It should be called by main, only once, before anything else starts. Once
// the configuration is set it should be treated as immutable an not changed.
func Config(c config.Config) {
	cfg = pkgConfig{
		crowdSize: c.Inventory.CrowdSize,
	}
}

type state struct {
	actor *Thing
	buf   map[*Thing]*strings.Builder
	cmd   string
	input string
	word  []string
}

// stopWords is a lookup table of words that can be removed from parsed input.
var stopWords = func() map[string]struct{} {
	m := make(map[string]struct{})
	for _, word := range []string{
		"A", "AN", "FROM", "IN", "INTO", "OF", "OUT", "SOME", "THE", "TO", "WITH",
	} {
		m[word] = struct{}{}
	}
	return m
}()

var newline = []byte("\n")

func NewState(t *Thing) *state {
	return &state{actor: t, buf: make(map[*Thing]*strings.Builder)}
}

func (s *state) Parse(input string) (cmd string) {
	if input = strings.TrimSpace(input); len(input) != 0 {
		// Stop the world for everyone else...
		BWL.Lock()
		defer BWL.Unlock()

		s.parse(input)
		s.mailman()
	}
	return s.cmd
}

func (s *state) parse(input string) {
	s.word = strings.Fields(strings.ToUpper(input))

	// Simple stop word removal
	keep := s.word[:0]
	for _, word := range s.word {
		if _, found := stopWords[word]; !found {
			keep = append(keep, word)
		}
	}
	s.word = keep

	// Nothing left?
	if len(s.word) == 0 {
		s.Msg(s.actor, "Eh?")
		return
	}

	s.cmd, s.word = s.word[0], s.word[1:]
	s.input = strings.TrimSpace(input[len(s.cmd):])

	if handler, ok := commandHandlers[s.cmd]; ok {
		savedDA := s.actor.As[DynamicAlias]
		s.actor.As[DynamicAlias] = "SELF"
		handler(s)
		if s.actor.Is&Freed != Freed {
			s.actor.As[DynamicAlias] = savedDA
		}
	} else {
		s.Msg(s.actor, "Eh?")
	}
}

// subparse parses new input reusing the current actor and buffers from the
// current state. This is useful for commands that want to be able to take
// advantage of the functionality other commands.
func (s *state) subparse(input string) {
	s2 := &state{actor: s.actor, buf: s.buf}
	s2.parse(input)
}

// mailman delivers queued messages to player's mailboxes. Messages can be
// queued for a specific player or for a location. If queued for a location,
// messages will be sent to all player at the location - unless they have
// received a specific message. Messages sent to the actor are always priority
// messages, others are not. See mailbox.Send for details of message priority.
//
// Note that even though commands are processed under the BWL mailboxes can be
// deleted at anytime due to network errors. This is not a problem, if the UID
// for a buffer is not for an existing mailbox or location it will be ignored
// and cleaned up.
func (s *state) mailman() {

	for ref, buf := range s.buf {
		// Send to specific players - race between Exists & Send is okay
		if mailbox.Exists(ref.As[UID]) {
			mailbox.Send(ref.As[UID], ref == s.actor, buf.String())
			continue
		}
		// Send to players at location, omitting players that are receiving
		// specific messages.
		for uid, who := range ref.Who {
			if s.buf[who] == nil {
				mailbox.Send(uid, false, buf.String())
			}
		}
	}

	// Cleanup buffers
	for ref, buf := range s.buf {
		buf.Reset()
		delete(s.buf, ref)
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
	if s.buf[recipient] == nil {
		s.buf[recipient] = &strings.Builder{}
		if recipient != s.actor {
			s.buf[recipient].Write(newline)
		}
	} else {
		s.buf[recipient].Write(newline)
	}
	for _, t := range text {
		s.buf[recipient].WriteString(t)
	}
}

// MsgAppend works the same as Msg, but does not force a line-feed to be added
// before appending the text. This can be used to build messages a piece at a
// time. It is safe to call MsgAppend for a recipient, even if Msg has not been
// called first.
func (s *state) MsgAppend(recipient *Thing, text ...string) {
	if s.buf[recipient] == nil {
		s.Msg(recipient, text...)
		return
	}
	for _, t := range text {
		s.buf[recipient].WriteString(t)
	}
}

// Prompt sets the player's current prompt to the given text.
//
// NOTE: A future update will make the prompt configurable. The prompt will
// also be  dynamic showing the players statistics.
func (s *state) Prompt(text string) {
	mailbox.Suffix(s.actor.As[UID], text)
}
