// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/mailbox"
	"code.wolfmud.org/WolfMUD.git/text"
)

// World contains all of the top level locations for the current game world.
// WorldStart only contains valid player starting locations. Both are protected
// by the BWL (Big World Lock).
var (
	BWL        sync.Mutex
	WorldStart []*Thing       // Starting locations
	World      = make(Things) // All top level locations
	Players    = make(Things) // Current in-game players
)

type pkgConfig struct {
	crowdSize   int // Represents minimum number of players considered a crowd
	debugThings bool
	debugEvents bool
	playerPath  string
}

// cfg setup by Config and should be treated as immutable and not changed.
var cfg pkgConfig

// Config sets up package configuration for settings that can't be constants.
// It should be called by main, only once, before anything else starts. Once
// the configuration is set it should be treated as immutable an not changed.
func Config(c config.Config) {
	cfg = pkgConfig{
		crowdSize:   c.Inventory.CrowdSize,
		debugThings: c.Debug.Things,
		debugEvents: c.Debug.Events,
		playerPath:  filepath.Join(c.Server.DataPath, "players"),
	}
}

type state struct {
	actor   *Thing
	buf     map[*Thing]*strings.Builder
	cmd     string
	input   string
	history [3]string
	word    []string
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

const (
	noScripting   = false
	withScripting = true
)

// Parse allows commands to be executed, from outside the package, with
// scripting disabled.
func (s *state) Parse(input string) (cmd string) {
	recall := false
	if input == "!" || input == "!!" || input == "!!!" {
		input = s.history[len(input)-1]
		recall = true
	}

	// If input isn't empty transfer it to the output area
	if input = strings.TrimSpace(input); len(input) > 0 {
		s.Msg(s.actor, text.Prompt, ">", input, text.Reset)
	}

	cmd = s.preParse(input, noScripting)

	if !recall && cmd != "/!" && cmd != "/HISTORY" && input != s.history[0] {
		copy(s.history[1:], s.history[0:])
		s.history[0] = input
	}

	return
}

// Script allows commands to be executed, from outside the package, with
// scripting enabled.
func (s *state) Script(input string) (cmd string) {
	return s.preParse(input, withScripting)
}

func (s *state) preParse(input string, allowScripting bool) (cmd string) {
	if input = strings.TrimSpace(input); len(input) == 0 {
		return ""
	}

	// Stop the world for everyone else...
	BWL.Lock()
	defer BWL.Unlock()

	s.parse(input, allowScripting)
	s.mailman()

	return s.cmd
}

func (s *state) parse(input string, allowScripting bool) {
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

	if !allowScripting && s.cmd[0] == '$' {
		s.Msg(s.actor, "Eh?")
		return
	}

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
	s2.parse(input, withScripting)
}

// subparseFor parses new input for an alternative actor reusing the current
// buffers from the current state. This is useful for commands that want to
// cause an alternative actor to perform a command. For example a GIVE command
// could be implemented as the actor performing a DROP and the receiver
// performing a GET.
//
// NOTE: The performed command is not passed back to the alternative actor's
// client. This means, for example, the HIT command cannot use the QUIT command
// if the alternative actor is killed - the client code will not see the QUIT.
func (s *state) subparseFor(actor *Thing, input string) {

	// 'mark' messages already sent to the original actor and current location
	markA := s.buf[s.actor].Len()
	markL := s.buf[s.actor.Ref[Where]].Len()

	s2 := &state{actor: actor, buf: s.buf}
	s2.parse(input, withScripting)

	// If the original actor already had messages and we have new location
	// messages, copy the additional location messages to the actor as they will
	// not be regarded as observers - as they already had specific messages.
	if markA != 0 && markL != s.buf[s.actor.Ref[Where]].Len() {
		s.buf[s.actor].WriteString(s.buf[s.actor.Ref[Where]].String()[markL:])
	}
}

// mailman delivers queued messages to player's mailboxes. Messages can be
// queued for a specific player or for a location. If queued for a location,
// messages will be sent to all players at the location - unless they have
// received a specific message. Messages sent to specific players are always
// priority messages, other messages are not priority. See mailbox.Send for
// details of message priority.
//
// Note that even though commands are processed under the BWL mailboxes can be
// deleted at anytime due to network errors. This is not a problem, if the UID
// for a buffer is not for an existing mailbox or location it will be ignored
// and cleaned up.
func (s *state) mailman() {

	for ref, buf := range s.buf {
		// Send to specific players - Exists/Send race ok, handled by mailbox
		if mailbox.Exists(ref.As[UID]) {
			mailbox.Send(ref.As[UID], true, buf.String())
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
	}
	s.buf[recipient].Write(newline)
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

// Log takes the same parameters as fmt.Printf and writes the resulting message
// to the log. The message will automatically be appended with the UID of the
// actor. For example:
//
//  [#UID-202] Quitting: 72de37d1b2be008b83e760ef74cc460a
//
func (s *state) Log(f string, a ...interface{}) {
	f = fmt.Sprintf("[%s] %s", s.actor.As[UID], f)
	log.Printf(f, a...)
}

// StatusUpdate updates the player's statistics on the status bar. The
// messages are sent as priority so that they do not effect the de-spamming of
// other messages.
func (s state) StatusUpdate(who *Thing) {
	mailbox.Send(who.As[UID], true, fmt.Sprintf(
		"%s Health: %[2]d/%[3]d\x1b8",
		who.As[StatusSeq], who.Int[HealthCurrent], who.Int[HealthMaximum],
	))
}
