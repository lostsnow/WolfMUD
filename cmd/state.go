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
	bytes.Buffer
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
	input       []string      // The original input of the actor
	cmd         string        // The current command being processed
	words       []string      // Input split into uppercased words
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

	s.input = strings.Fields(input)

	// BUG: Stop words are currently experimental. Use of removeStopWords means
	// that s.input is no longer the original input anymore, but s.input and
	// s.words do still match up.
	s.input = internal.RemoveStopWords(s.input)

	s.words = make([]string, len(s.input))
	for x, o := range s.input {
		s.words[x] = strings.ToUpper(o)
	}

	// Make sure we don't try to index beyond the
	// number of words we have and cause a panic
	switch l := len(s.words); {
	case l > 1:
		s.cmd, s.words = s.words[0], s.words[1:]
		s.input = s.input[1:]
	case l > 0:
		s.cmd, s.words = s.words[0], []string{}
		s.input = []string{}
	}

	// Need to determine the actor's current location so we can lock it. As
	// commands frequently need to know the current location also, we stash it in
	// the state for later reuse.
	if s.where = attr.FindLocate(t).Where(); s.where != nil {
		s.AddLock(s.where)
	}

	return s
}

// parse repeatedly calls sync until the list of locks after the call to sync
// is the same as before the call to sync. sync is always called at least once.
//
// When sync calls the dispatcher to handle commands the command may determine
// it needs to hold additional locks. In this case the command calls AddLock
// for each additional lock it needs and simply returns. parse will detect that
// the list of locks has changed and call sync again, this time with the new
// list of locks.
//
// NOTE: There is usually at least one lock, added by NewState, which is the
// containing Inventory of the current actor - if it is locatable.
//
// NOTE: At the moment locks are only added - using AddLock. A change in the
// lock list can therefore be detected by simply checking the length of the
// list. If at a later time we need to be able to remove locks as well this
// simple length check will not be sufficient.
func (s *state) parse(dispatcher func(*state)) {
	for l := -1; l != 0; {
		l = len(s.locks)
		s.sync(dispatcher)
		l -= len(s.locks)
	}
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

// silent allows a command to be processed without sending messages to specific
// targets. The passed actor, participant and observers flags can be set to
// prevent messages from being sent to specific targets.
//
// TODO: This is a simple but not a very efficient way to implement this as the
// message are still 'sent' and we just chop them off again by truncating the
// buffers. Ideally we should stop the buffers from being written to in the
// first place.
//
// BUG(diddymus): We don't treat observer differently to observers - should we?
func (s *state) silent(actor, participant, observers bool, cmd func(*state)) {

	// If no flags set we can just process the command normally...
	if !actor && !participant && !observers {
		cmd(s)
		return
	}

	var (
		aMark int
		pMark int
		oMark map[has.Inventory]int
	)

	// Mark the current length of the buffers we want to silence
	if actor {
		aMark = s.msg.actor.Len()
	}
	if participant {
		pMark = s.msg.participant.Len()
	}
	if observers {
		oMark = make(map[has.Inventory]int, len(s.msg.observers))
		for k, observer := range s.msg.observers {
			oMark[k] = observer.Len()
		}
	}

	cmd(s)

	// Truncate the buffers back to their marked length for buffers we silenced
	if actor && aMark != s.msg.actor.Len() {
		s.msg.actor.Truncate(aMark)
	}
	if participant && pMark != s.msg.participant.Len() {
		s.msg.participant.Truncate(pMark)
	}
	if observers {
		for k, observer := range s.msg.observers {
			if oMark[k] != observer.Len() {
				observer.Truncate(oMark[k])
			}
		}
	}
}

// sync is called by parse to do the actual locking and unlocking. Having this
// separate from parse takes advantage of unwinding the locks using defer. This
// makes both parse and sync very simple.
func (s *state) sync(dispatcher func(s *state)) {
	for _, l := range s.locks {
		l.Lock()
		defer l.Unlock()
	}

	s.allocateBuffers()
	l := len(s.locks)
	dispatcher(s)

	// If we don't add any new locks process any pending messages before we
	// release our locks
	if l-len(s.locks) == 0 {
		s.messenger()
	}
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
		s.msg.actor = &buffer{Buffer: *bytes.NewBuffer(make([]byte, 0, (80*24)/2))}
		s.msg.participant = &buffer{Buffer: *bytes.NewBuffer([]byte{})}
		s.msg.observers = make(map[has.Inventory]*buffer)
		s.msg.participant.WriteByte(byte('\n'))
	}

	for _, l := range s.locks {
		if _, ok := s.msg.observers[l]; !ok {
			s.msg.observers[l] = &buffer{Buffer: *bytes.NewBuffer([]byte{})}
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
