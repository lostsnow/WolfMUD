// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd/internal"
	"code.wolfmud.org/WolfMUD.git/has"

	"bytes"
	"strings"
)

// buffer is our extended version of a bytes.Buffer so that we can add some
// convience methods.
type buffer struct {
	*bytes.Buffer
}

// WriteJoin takes a number of strings and writes them into the buffer. It's a
// convenience method to save writing multiple WriteString statements and an
// alternative to additional allocations due to concatenation.
//
// The return value n is the total length of all s, in bytes; err is always nil.
// The underlying bytes.Buffer may panic if it becomes too large.
func (b *buffer) WriteJoin(s ...string) (n int, err error) {
	for _, s := range s {
		x, _ := b.WriteString(s)
		n += x
	}
	return n, nil
}

// state contains the current parsing state for commands. The state fields may
// be modified directly except for locks. The AddLocks method should be used to
// add locks, CanLock can be called to see if a lock has already been added.
//
// NOTE: where is only set when the state is created. If the actor moves to
// another location where should be updated as well.
//
// TODO: Need to document msg buffers properly
type state struct {
	actor       has.Thing     // The Thing executing the command
	where       has.Inventory // Where the actor currently is
	participant has.Thing     // The other Thing participating in the command
	input       []string      // The original input of the actor minus cmd
	cmd         string        // The current command being processed
	words       []string      // Input as uppercased words, less stopwords
	ok          bool          // Flag to indicate if command was successful

	// DO NOT MANIPULATE LOCKS DIRECTLY - use AddLock and see it's comments
	locks []has.Inventory // List of locks we want to be holding

	// msg is a collection of buffers for gathering messages to send back as a
	// result of processing a command. Note observer is setup as an 'alias' for
	// observers[s.where] - observer and observers[s.where] point to the same
	// buffer.
	msg struct {
		actor       *buffer
		participant *buffer
		observer    *buffer
		observers   map[has.Inventory]*buffer
	}
}

// NewState returns a *state initialised with the passed Thing and input. If
// the passed Thing is locatable the containing Inventory is added to the lock
// list, but the lock is not taken at this point.
func NewState(t has.Thing, input string) *state {

	s := &state{
		actor: t,
		locks: make([]has.Inventory, 0, 2), // Common case is only 1 or 2 locks
	}

	s.tokenizeInput(input)

	// Need to determine the actor's current location so we can lock it. As
	// commands frequently need to know the current location also, we stash it in
	// the state for later reuse.
	s.where = attr.FindLocate(t).Where()
	s.AddLock(s.where)

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

		for x, o := range s.words {
			s.words[x] = strings.ToUpper(o)
		}

		s.cmd, s.words = s.words[0], s.words[1:]
		s.input = s.input[1:]
	}
}

// parse repeatedly calls sync until it returns true.
//
// When sync handles a command the command may determine it needs to hold
// additional locks. In this case sync will return false and should be called
// again. This repeats until the list of locks is complete, the command
// processed and sync returns true.
func (s *state) parse() {
	for !s.sync() {
	}
}

// sync is called to do the actual locking/unlocking for commands. Having this
// separate from takes advantage of unwinding the locks using defer. This makes
// sync very simple. If the list of locks before and after handling a command
// are the same we are 'in sync' and had all the locks we needed to process the
// command. In this case we return true. If more locks need to be acquired we
// return false and should be called again.
//
// NOTE: There is usually at least one lock, added by NewState, which is the
// containing Inventory of the current actor - if it is locatable.
//
// NOTE: At the moment locks are only added - using AddLock. A change in the
// lock list can therefore be detected by simply checking the length of the
// list. If at a later time we need to be able to remove locks as well this
// simple length check will not be sufficient.
func (s *state) sync() (inSync bool) {
	for _, l := range s.locks {
		l.Lock()
		defer l.Unlock()
	}

	s.allocateBuffers()
	l := len(s.locks)

	s.handleCommand()

	// If we don't add any new locks we are 'in sync'. Therefore set inSync flag
	// and process any pending messages before all of the locks get released.
	if l-len(s.locks) == 0 {
		inSync = true
		s.messenger()
	}
	return
}

// handleCommand runs the registered handler for the current state command. If
// a handler cannot be found a message will be written to the actor's output
// buffer.
//
// BUG(diddymus): Should this be moved into handler.go?
func (s *state) handleCommand() {
	switch handler, valid := handlers[s.cmd]; {
	case valid:
		handler(s)
	default:
		s.msg.actor.WriteString("Eh?")
	}
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
// TODO: Suppression of messages is not very efficient as any messages are
// still 'sent' and we just chop them off again by truncating the buffers.
// Ideally we should stop the buffers from being written to in the first place.
//
// BUG(diddymus): We don't treat observer differently to observers - should we?
func (s *state) script(actor, participant, observers bool, inputs ...string) {

	input := strings.Join(inputs, " ")

	i, w, c := s.input, s.words, s.cmd // Save state

	a := s.msg.actor.Len()
	p := s.msg.participant.Len()
	o := make(map[has.Inventory]int)
	for where, observer := range s.msg.observers {
		o[where] = observer.Len()
	}

	if a > 0 {
		s.msg.actor.WriteString("\n")
	}
	if p > 0 {
		s.msg.participant.WriteString("\n")
	}
	for where, o := range o {
		if o > 0 {
			s.msg.observers[where].WriteString("\n")
		}
	}

	s.tokenizeInput(input)
	s.ok = false
	s.handleCommand()

	// If anything is written to the buffers during processing:
	// 	- append a newline to the buffer if output not suppressed. This will
	// 		cause each action to start on it's own line.
	// 	- if messages are suppressed for a buffer truncate back to initial
	// 		length.
	if l := s.msg.actor.Len(); actor && (l-a > 1) {
		a = l
	}
	s.msg.actor.Truncate(a)

	if l := s.msg.participant.Len(); participant && (l-p > 1) {
		p = l
	}
	s.msg.participant.Truncate(p)

	for where, observer := range s.msg.observers {
		if l := observer.Len(); observers && (l-o[where] > 1) {
			o[where] = l
		}
		observer.Truncate(o[where])
	}

	s.input, s.words, s.cmd = i, w, c // Restore state
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

// scriptAll is a helper method that is equivalent to calling script with
// messages suppressed for any participant or observers. Only the actor will
// receive any messages.
func (s *state) scriptActor(input ...string) {
	s.script(true, false, false, input...)
}

// messenger is used to send buffered messages to the actor, participant and
// observers. The participant may be in another location to the actor - such as
// when throwing something at someone or shooting someone.
//
// For the actor we don't check the buffer length to see if there is anything
// in it to send. We always send to the actor so that we can redisplay the
// prompt even if they just hit enter.
//
// NOTE: Messages are not broadcast to observers in a crowded location.
func (s *state) messenger() {

	if s.actor != nil {
		attr.FindPlayer(s.actor).Write(s.msg.actor.Bytes())
	}

	if s.participant != nil && s.msg.participant.Len() > 1 {
		attr.FindPlayer(s.participant).Write(s.msg.participant.Bytes())
	}

	if len(s.msg.observers) == 0 || s.where == nil {
		return
	}

	for where, buffer := range s.msg.observers {
		if where.Crowded() || buffer.Len() == 1 {
			continue
		}
		msg := buffer.Bytes()
		for _, c := range where.Contents() {
			if c != s.actor && c != s.participant {
				attr.FindPlayer(c).Write(msg)
			}
		}
	}

	s.deallocateBuffers()
}

// CanLock returns true if the specified Inventory is in the list of locks and
// could be locked, otherwise false. It does NOT determine if the lock is
// currently held or not.
func (s *state) CanLock(i has.Inventory) bool {
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
// This method uses a version of an online straight insertion sort. For the
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

	if i == nil || s.CanLock(i) {
		return
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

// allocateBuffers sets up the message buffers for the actor, participant and
// observers. The participant and observers buffers need an initial linefeed to
// move the cursor off of the client's prompt line - for the actor this is done
// when they hit enter. The actor's buffer is initially set to half a page
// (half of 80 columns by 24 lines) as it is common to be sending location
// descriptions back to the actor. Half a page is arbitrary but seems to be
// reasonable.
func (s *state) allocateBuffers() {
	if s.msg.actor == nil {
		s.msg.actor = &buffer{Buffer: bytes.NewBuffer(make([]byte, 0, (80*24)/2))}
		s.msg.participant = &buffer{Buffer: bytes.NewBuffer([]byte{})}
		s.msg.observers = make(map[has.Inventory]*buffer)
		s.msg.participant.WriteByte(byte('\n'))
	}

	for _, l := range s.locks {
		if _, ok := s.msg.observers[l]; !ok {
			s.msg.observers[l] = &buffer{Buffer: bytes.NewBuffer([]byte{})}
			s.msg.observers[l].WriteByte('\n')
		}
	}
	s.msg.observer = s.msg.observers[s.where]
}

// deallocateBuffers releases the references to message buffers for the actor,
// participant and observers.
func (s *state) deallocateBuffers() {
	s.msg.actor = nil
	s.msg.participant = nil
	s.msg.observer = nil
	for x := range s.msg.observers {
		s.msg.observers[x] = nil
		delete(s.msg.observers, x)
	}
}
