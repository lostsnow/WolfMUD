// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: $CLEANUP item
func init() {
	AddHandler(Cleanup, "$cleanup")
}

func Cleanup(s *state) {

	// Do we have item to cleanup specified on command?
	if len(s.words) == 0 {
		return
	}

	// Search for item to perform action.
	alias := s.words[0]
	what := s.where.Search(alias)

	// If item not found all we can do is bail.
	if what == nil {
		return
	}

	oc := attr.FindOnCleanup(what)
	msg := oc.CleanupText()

	// Clean up will not be seen if we specifically have an empty message. It
	// will also not be seen if there are no players here to see it or it's too
	// crowded. In these cases just junk Thing.
	if (oc.Found() && msg == "") || !s.where.Players() || s.where.Crowded() {
		s.scriptNone("junk", alias)

		// s.ok will be that of the scripted JUNK and we will also retry if JUNK
		// modifies the locks
		return
	}

	// Clean up will be seen so add default message if we don't have one
	if !oc.Found() {
		name := attr.FindName(what).Name("something")
		msg = "You are sure you noticed " + name + " here, but you can't see it now."
	}

	// Script JUNK of item to remove it and relock/abort if needed
	l := len(s.locks)
	s.scriptNone("junk", alias)
	if l != len(s.locks) || !s.ok {
		return
	}

	// Display messages. Only notify the actor if it's not the Thing issuing the
	// command.
	if s.actor.UID() != what.UID() {
		s.msg.Actor.SendInfo(msg)
	}
	s.msg.Observer.SendInfo(msg)

	s.ok = true
}
