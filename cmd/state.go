// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"io"
	"strings"
	"time"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd/internal"
	"code.wolfmud.org/WolfMUD.git/event"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/message"
	"code.wolfmud.org/WolfMUD.git/stats"
)

func init() {
	event.Script = Script
}

// state contains the current parsing state for commands. The state fields may
// be modified directly except for locks. The AddLocks method should be used to
// add locks, CanLock can be called to see if a lock has already been added.
//
// Care should be taken if the state.where field is updated by a command.
// Updating the state.where field to an Inventory not covered by a lock, via
// AddLock, will cause the lock list to be cleared and processing for the
// command will start over. This is due to the fact that state.sync will detect
// the actor has moved and that the locks are now invalid - as state.were is
// not covered by a lock. See cmd.move for an example of updating state.where
// with command.
//
// TODO: Need to document msg buffers properly
type state struct {
	actor       has.Thing     // The Thing executing the command
	where       has.Inventory // Where the actor currently is
	participant has.Thing     // The other Thing participating in the command
	input       []string      // The original input of the actor minus cmd
	cmd         string        // The current command being processed
	words       []string      // Input as uppercased words, less cmd & stopwords
	ok          bool          // Flag to indicate if command was successful
	scripting   bool          // Is state in scripting mode?

	// locks is the list of locks we want to be holding. locks has a specific
	// ordering and should be added using AddLock, see it's comments.
	locks []has.Inventory

	// msg contains the message buffers for sending data to different recipients
	msg message.Msg
}

// Parse initiates processing of the input string for the specified Thing. The
// input string is expected to be input from a player.
//
// Parse runs with state.scripting set to false, disallowing scripting specific
// commands from being executed by players directly.
//
// When sync handles a command the command may determine it needs to hold
// additional locks. In this case sync will return false and should be called
// again. This repeats until the list of locks are acquired, the command
// processed and sync returns true.
func Parse(t has.Thing, input string) {
	s := newState(t, input)
	for !s.sync() {
	}
}

// Script processes the input string the same as Parse. However Script runs
// with the state.scripting flag set to true, permitting scripting specific
// commands to be executed.
func Script(t has.Thing, input string) string {
	s := newState(t, input)
	s.scripting = true
	for !s.sync() {
	}
	return s.cmd
}

// newState returns a *state initialised with the passed Thing and input.
func newState(t has.Thing, input string) *state {

	s := &state{
		actor: t,
		locks: make([]has.Inventory, 0, 2), // Common case is only 1 or 2 locks
	}

	s.tokenizeInput(input)

	return s
}

// tokenizeInput takes the given string and breaks it into uppercased words
// which are stored in the current state. After processing s.cmd will contain
// the leading command, uppercased. s.input will contain the original input
// minus the leading s.cmd. s.words will contain the input, uppercased with
// stopwords and the leading s.cmd removed. For example:
//
//	input = "Say I'm in need of help!"
//	s.cmd = "SAY"
//	s.input = []string{"I'm", "in", "need", "of", "help!"}
//	s.words = []string{"I'M", "NEED", "HELP!"}
//
func (s *state) tokenizeInput(input string) {
	s.input = strings.Fields(input)

	if len(s.input) > 0 {
		s.words = internal.RemoveStopWords(s.input)

		// If the input only consists of stop words fake the "Eh?" command being
		// entered to get an "Eh?" response from dispatchHandler - "Eh?" is not a
		// valid command and seems an appropriate choice
		if len(s.words) == 0 {
			s.cmd = "Eh?"
			return
		}

		for x, o := range s.words {
			s.words[x] = strings.ToUpper(o)
		}

		s.cmd, s.words = s.words[0], s.words[1:]
		s.input = s.input[1:]
	}
}

// sync acquires locks reuired for command processing. If the number of locks
// before and after handling a command are the same we are 'in sync' and had
// all the locks we needed to process the command and return true. If a command
// required, and added via state.AddLock, additional locks then return false
// and sync should be called again.
func (s *state) sync() (inSync bool) {
	lockStart := time.Now()

	for _, l := range s.locks {
		l.Lock()
		defer l.Unlock()
	}

	lockWait := <-stats.MaxLockWait
	if diff := time.Now().Sub(lockStart); diff > lockWait {
		lockWait = diff
	}
	stats.MaxLockWait <- lockWait

	// If actor not where we think it is s.where and s.locks will be invalid, and
	// we will be acquiring the wrong locks, so start over. On our first pass
	// s.where will never match so this performs initialisation as well.
	//
	// Note: Just checking s.where is not enough. Consider a location L, with a
	// container C and actor A. The actor is in the container which it at the
	// location. Our Inventory hierarchy is L←C←A, s.where points to C and we
	// lock L. If C moves to location L' our Inventory hierarchy is now L'←C←A
	// but s.where is unchanged and still points to C but now the lock needs to
	// be on L' and not L.
	if l := attr.FindLocate(s.actor).Where(); s.where != l || !s.CanLock(l) {
		s.where = l
		for x := range s.locks {
			s.locks[x] = nil
		}
		s.locks = s.locks[:0]
		s.AddLock(l)
		return false
	}

	// If final location of actor is nowhere and actor is not a location then
	// nothing to process. Locations themselves are always nowhere - they are the
	// top of the Inventory tree, but they can process events like resets and
	// clean-ups.
	if s.where == nil && !attr.FindExits(s.actor).Found() {
		return true
	}

	s.msg.Allocate(s.where, s.locks)
	l := len(s.locks)

	dispatchHandler(s)

	// If we don't add any new locks we are 'in sync'. Therefore set inSync flag
	// and process any pending messages before all of the locks get released.
	if l-len(s.locks) == 0 {
		inSync = true
		s.messenger()
	}
	return
}

// script executes the given input as a command using the current state.
// Messages are only sent to the actor, participant and any observers if the
// relevant actor, participant or observers parameter is set to true. The
// script method should only be called by commands that want to execute
// sub-commands. For example a 'GIVE ITEM' command might be implemented by
// scripting together a 'DROP ITEM' and a 'GET ITEM' - thereby reusing the code
// implementing the DROP and GET commands.
//
// The command we are scripting will be processed with the current state,
// including any currently held locks and any current message buffers. The
// value of s.ok will be the result of the scripted command.
//
// For convenience, and to avoid concatenation when building commands to be
// scripted, the input string can be passed in as multiple strings that will
// joined together automatically with space separators.
//
// script will automatically set and restore the state.scripting flag allowing
// scripting specific commands to be executed.
//
// TODO: Suppression of messages is not very efficient as any messages are
// still 'sent' and we just chop them off again by truncating the buffers.
// Ideally we should stop the buffers from being written to in the first place.
//
// BUG(diddymus): We don't treat observer differently to observers - should we?
func (s *state) script(actor, participant, observers bool, inputs ...string) {

	input := strings.Join(inputs, " ")

	i, w, c, sc := s.input, s.words, s.cmd, s.scripting // Save state

	// Set silent mode on buffers storing old modes
	a := s.msg.Actor.Silent(!actor)
	p := s.msg.Participant.Silent(!participant)
	ot, of := s.msg.Observers.Silent(!observers)

	s.tokenizeInput(input)
	s.ok = false
	s.scripting = true
	dispatchHandler(s)

	// Restore old silent modes
	s.msg.Actor.Silent(a)
	s.msg.Participant.Silent(p)
	ot.Silent(true)
	of.Silent(false)

	s.input, s.words, s.cmd, s.scripting = i, w, c, sc // Restore state
}

// scriptAll is a helper method that is equivalent to calling script with no
// messages suppressed for the actor, participant or observers.
func (s *state) scriptAll(input ...string) {
	s.script(true, true, true, input...)
}

// scriptAll is a helper method that is equivalent to calling script with all
// messages suppressed.
func (s *state) scriptNone(input ...string) {
	s.script(false, false, false, input...)
}

// scriptActor is a helper method that is equivalent to calling script with
// messages suppressed for any participant or observers. Only the actor will
// receive any messages.
func (s *state) scriptActor(input ...string) {
	s.script(true, false, false, input...)
}

// asParticipant executes the given input as a command for the participant
// using the current state. It is functionally equivalent to cmd.script but
// with the actor and participant roles temporarily reversed for the duration
// of the command. If the current participant is nil this method will just
// return. See cmd.script for additional information.
//
// BUG(diddymus): It's currently assumed the actor and participant are at the
// same location. This is important as s.were is not updated when roles are
// switched, swapping locations could have implications for the locks being
// held. More investigation and testing required.
func (s *state) asParticipant(inputs ...string) {

	if s.participant == nil {
		return
	}

	// Reverse actor and participant roles
	s.actor, s.participant = s.participant, s.actor
	s.msg.Actor, s.msg.Participant = s.msg.Participant, s.msg.Actor

	s.script(true, true, true, inputs...)

	// Restore actor and participant roles
	s.actor, s.participant = s.participant, s.actor
	s.msg.Actor, s.msg.Participant = s.msg.Participant, s.msg.Actor

}

// messenger is used to send buffered messages to the actor, participant and
// observers. The participant may be in another location to the actor - such as
// when throwing something at someone or shooting someone.
//
// For the actor we don't check the buffer length to see if there is anything
// in it to send. We always send to the actor so that we can redisplay the
// prompt even if they just hit enter.
func (s *state) messenger() {

	var p has.Player

	if s.actor != nil {
		if p = attr.FindPlayer(s.actor); p.Found() {
			s.msg.Actor.Deliver(p)
		}
	}

	if s.participant != nil && s.msg.Participant.Len() > 0 {
		if p = attr.FindPlayer(s.participant); p.Found() {
			s.msg.Participant.Deliver(p)
		}
	}

	for where, buffer := range s.msg.Observers {
		if buffer.Len() == 0 {
			continue
		}
		players := []io.Writer{}
		for _, p := range where.Players() {
			if p != s.actor && p != s.participant {
				players = append(players, attr.FindPlayer(p))
			}
		}
		buffer.Deliver(players...)
	}

	s.msg.Deallocate()
}

// CanLock returns true if the specified Inventory is covered in the list of
// locks and could be locked, otherwise false. It does NOT determine if the
// lock is currently held or not.
func (s *state) CanLock(i has.Inventory) bool {

	if i == nil {
		return true
	}

	i = i.Outermost()

	for _, l := range s.locks {
		if i == l {
			return true
		}
	}
	return false
}

// AddLock takes an Inventory and adds it to the lock list in the correct
// position relative to other Inventory in the list.
//
// Locks should always be acquired in lock ID sequence lowest to highest to
// avoid deadlocks. By using this method the lock list can easily be iterated
// via a range and in the correct sequence required.
//
// This method uses a version of an inline straight insertion sort. For the
// vast majority of cases we are only dealing with 1 or 2 locks. Actions in the
// same location like get, drop, examine, etc. only require 1 lock. Moving from
// one location to another location requires 2 locks. Having more that 2 locks
// is rare but could occure with things like area or line of sight effects.
//
// As we can broadcast messages to anyone in any of the locked locations we
// also setup an observers message buffer for each added lock. The message
// buffers can then be accessed using:
//
//	s.msg.observers[i]
//
// where i is a location's Inventory.
//
// NOTE: We cannot add the same lock twice otherwise we would deadlock
// ourselves when locking - currently we silently drop duplicate locks.
func (s *state) AddLock(i has.Inventory) {

	if i == nil || !i.Found() {
		return
	}

	i = i.Outermost()

	for _, l := range s.locks {
		if i == l {
			return
		}
	}

	s.locks = append(s.locks, i)
	l := len(s.locks)

	if l == 1 {
		return
	}

	u := i.LockID()
	for x := 0; x < l; x++ {
		if s.locks[x].LockID() > u {
			copy(s.locks[x+1:l], s.locks[x:l-1])
			s.locks[x] = i
			break
		}
	}
}
