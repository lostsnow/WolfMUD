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
type Interface interface {
	Process(*Command) (handled bool)
	//IsLocked(thing thing.Interface) bool
}

// Command represents the state of the command currently being processed.
type Command struct {
	Issuer thing.Interface // What is issuing the command
	Verb   string          // 1st word (verb): GET, DROP, EXAMINE etc
	Nouns  []string        // 2nd...nth words
	Target *string         // Alias for 2nd word - normally the verb's target
	Locks  []thing.Interface
	Relock thing.Interface
}

// New creates a new Command instance. The string is broken into words using
// whitespace as the separator. The first word is usually the verb - get,
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

func (c *Command) IsLocked(thing thing.Interface) bool {
	a := thing.UniqueId()
	for _, b := range c.Locks {
		if a == b.UniqueId() {
			return true
		}
	}
	return false
}

func (c *Command) ClearRelock() {
	c.Relock = nil
}

// BUG(D) Should really implement sort interface?
func (c *Command) AddLock() {
	defer c.ClearRelock()

	if c.Relock != nil {
		c.Locks = append(c.Locks, c.Relock)
		if len(c.Locks) == 1 {
			return
		}
		for swap := true; swap; {
			swap = false
			for i := 0; i < len(c.Locks)-1; i++ {
				if i < len(c.Locks) {
					if c.Locks[i].UniqueId() > c.Locks[i+1].UniqueId() {
						c.Locks[i], c.Locks[i+1] = c.Locks[i+1], c.Locks[i]
						swap = true
					}
				}
			}
		}
	}
}
