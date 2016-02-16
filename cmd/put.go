// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
)

// Syntax: PUT item container
func init() {
	AddHandler(Put, "PUT")
}

func Put(s *state) {

	if len(s.words) == 0 {
		s.msg.actor.WriteString("You go to put something into something else...")
		return
	}

	tName := s.words[0]

	// Search ourselves for item to put into container
	tWhere := attr.FindInventory(s.actor)
	tWhat := tWhere.Search(tName)

	if tWhat == nil {
		s.msg.actor.WriteJoin("You have no '", tName, "' to put into anything.")
		return
	}

	// Get item's proper name
	tName = attr.FindName(tWhat).Name(tName)

	// Check a container was specified
	if len(s.words) < 2 {
		s.msg.actor.WriteJoin("What did you want to put ", tName, " into?")
		return
	}

	cName := s.words[1]

	// Search ourselves for container to put something into
	cWhat := tWhere.Search(cName)

	// If container not found and we are not somewhere we're not going to find it
	if cWhat == nil && s.where == nil {
		s.msg.actor.WriteJoin("There is no '", cName, "' to put ", tName, " into.")
		return
	}

	// If container not found search the inventory where we are
	if cWhat == nil {
		cWhat = s.where.Search(cName)
	}

	// Was container found?
	if cWhat == nil {
		s.msg.actor.WriteJoin("You see no '", cName, "' to put ", tName, " into.")
		return
	}

	// Unless our name is Klein we can't put something inside itself! ;)
	if tWhat == cWhat {
		s.msg.actor.WriteJoin("You can't put ", tName, " inside itself!")
		return
	}

	// Get container's proper name
	cName = attr.FindName(cWhat).Name(cName)

	// Check container is actually a container with an inventory
	cInv := attr.FindInventory(cWhat)
	if !cInv.Found() {
		s.msg.actor.WriteJoin("You cannot put ", tName, " into ", cName, ".")
		return
	}

	// Check for veto on item being put into container
	if veto := attr.FindVetoes(tWhat).Check("DROP", "PUT"); veto != nil {
		s.msg.actor.WriteString(veto.Message())
		return
	}

	// Check for veto on container
	if veto := attr.FindVetoes(cWhat).Check("PUT"); veto != nil {
		s.msg.actor.WriteString(veto.Message())
		return
	}

	// Remove item from where it is
	if tWhere.Remove(tWhat) == nil {
		s.msg.actor.WriteJoin("Something stops you putting ", tName, " anywhere.")
		return
	}

	// Put item into comtainer
	cInv.Add(tWhat)

	s.msg.actor.WriteJoin("You put ", tName, " into ", cName, ".")
	s.ok = true
}
