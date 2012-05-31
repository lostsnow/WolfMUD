// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package command implements the representation and state of a command that is
// being processed.
package command

import (
	"strings"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/responder"
)

// Interface should be implemented by anything that wants to process/react
// to commands. These commands are usually from players and mobiles but also
// commonly from other objects. For example a lever when pulled may issue an
// 'OPEN DOOR' command to get the door associated with the lever to open.
//
// The Process method when called should return true if the command was
// processed by the Thing implementing Process. Note that handled means the
// command was dealt with by a Thing. The outcome may be a success or a failure
// - but it WAS still handled.
//
// TODO: Beef up description when examples available.
// TODO: Need to document locking
type Interface interface {
	Process(*Command) (handled bool)
}

// Command represents the state of the command currently being processed.
// Command is also used to pass around locking information as the command is
// processed.
type Command struct {
	Issuer        thing.Interface   // What is issuing the command
	Verb          string            // 1st word (verb): GET, DROP, EXAMINE etc
	Nouns         []string          // 2nd...nth words
	Target        *string           // Alias for 2nd word - normally the verb's target
	Locks         []thing.Interface // Locks we want to hold
	locksModified bool              // Locks modified since last LocksModified() call?
}

// New creates a new Command instance. The input string is broken into words
// using whitespace as the separator. The first word is usually the verb - get,
// drop, examine - and the second word the target noun to act on - get ball,
// drop ball, examine ball. As this is a common case the second word cam also
// referenced via the alias Target.
func New(issuer thing.Interface, input string) *Command {
	words := strings.Split(strings.ToUpper(input), ` `)

	cmd := Command{}

	cmd.Issuer = issuer
	cmd.Verb = words[0]
	cmd.Nouns = words[1:]

	if len(words) > 1 {
		cmd.Target = &words[1]
	}

	return &cmd
}

// Respond implements the responder Interface. It is a quick shorthand for
// responding to the Thing that is issuing the command without having to do any
// additional bookkeeping.
//
// TODO: Need to also implement the broadcast Interface so we can just as easily
// respond to everyone present not issuing the command. As in:
//
//	c.Respond("You sneeze")
//
//	c.Broadcast("You see %s sneeze.", c.Issuer.Name())
//
// However to do this we need a location which Thing does not carry but will be
// implemented in a general 'object' which is a thing with location.
func (c *Command) Respond(format string, any ...interface{}) {
	if i, ok := c.Issuer.(responder.Interface); ok {
		i.Respond(format, any...)
	}
}

// CanLock checks if the command has the thing in it's locks list. This only
// determines if the thing is in the Locks slice - not if it is or is not
// actually locked.
func (c *Command) CanLock(thing thing.Interface) bool {
	for _, l := range c.Locks {
		if thing.IsAlso(l) {
			return true
		}
	}
	return false
}

// LocksModified returns true if the Locks slice has been modified since the
// Command was created or since the last call of LocksModified.
//
// NOTE: Calling this function will also reset the check to false.
func (c *Command) LocksModified() (modified bool) {
	modified = c.locksModified
	c.locksModified = false
	return
}

// AddLock takes a reference to a thing and adds it to the Locks slice in the
// correct position. Locks should always be acquired in unique Id sequence
// lowest to highest to avoid deadlocks. By using this method the Locks property
// can easily be iterated via a range and in the correct sequence required.
//
// NOTE: We cannot add the same Lock twice otherwise we would deadlock ourself
// when locking.
//
// This routine is a little cute and avoids doing any 'real' sorting to keep the
// elements in unique ID sequence. We add our lock to our slice. If we have one
// element only it's what we just added so we bail.
//
// If we have multiple elements we have the appended element on the end and need
// to check where it goes, shift the trailing element up by one the write our
// new element in:
//
//	3 7 9 4 <- append new element to end
//	3 7 9 4 <- correct place: 4 goes between 3 and 7
//	3 7 7 9 <- shift 7,9 up one overwriting our appended element
//	3 4 7 9 <- we now write our new element into our 'hole'
//
// What if we can't find an element with a unique Id greater than the one we are
// inserting?
//
//	3 7 9 10 <- append new element to end
//	3 7 9 10 <- correct place: 10 goes after 9, not insert point found
//	3 7 9 10 <- No shifting is done, appended element not over-written
//	3 4 7 10 <- new element already in correct place, nothing else to do
//
// This function could be more efficient with large numbers of elements by
// using a binary search to find the insertion point for the new element.
// However this would make the code more complex and we don't expect huge
// numbers of elements.
//
// Amazing, 347 characters of code, 1696 characters of comments!
func (c *Command) AddLock(t thing.Interface) {

	if t == nil || c.CanLock(t) {
		return
	}

	c.locksModified = true
	c.Locks = append(c.Locks, t)

	if l := len(c.Locks); l > 1 {
		uid := t.UniqueId()
		for i := 0; i < l; i++ {
			if uid > c.Locks[i].UniqueId() {
				copy(c.Locks[i+1:l], c.Locks[i:l-1])
				c.Locks[i] = t
				break
			}
		}
	}
}
