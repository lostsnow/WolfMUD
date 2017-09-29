// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/text"
)

// Syntax: TAKE item container
func init() {
	addHandler(take{}, "TAKE")
}

type take cmd

func (take) process(s *state) {

	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("You go to take something out of something else...")
		return
	}

	tName := s.words[0]

	// Was container specified? We have to check for the container first as the
	// item would be in the container, if there is no container specified we
	// cannot find the item and hence resolve the proper name for it.
	if len(s.words) < 2 {
		s.msg.Actor.SendBad("What did you want to take '", tName, "' out of?")
		return
	}

	cName := s.words[1]

	// Find the taking thing's own inventory. We remember this inventory as this
	// is where the item will be put if sucessfully taken from the container
	tWhere := attr.FindInventory(s.actor)

	// Search inventory for the container
	cWhat := tWhere.Search(cName)

	// If we have not foun the container and we are nowhere we are not going to
	// find the container so bail early
	if cWhat == nil && s.where == nil {
		s.msg.Actor.SendBad("You see no '", cName, "' to take anything from.")
		return
	}

	// If container not found yet search where we are
	if cWhat == nil {
		cWhat = s.where.Search(cName)
	}

	// Was container found?
	if cWhat == nil {
		s.msg.Actor.SendBad("You see no '", cName, "' to take things out of.")
		return
	}

	// Get container's proper name
	cName = attr.FindName(cWhat).Name(cName)

	// Check container is actually a container with an inventory
	cWhere := attr.FindInventory(cWhat)
	if !cWhere.Found() {
		s.msg.Actor.SendBad("You cannot take anything from ", cName)
		return
	}

	// Get actor's name
	who := attr.FindName(s.actor).Name("Someone")

	// Is item to be taken in the container?
	tWhat := cWhere.Search(tName)
	if tWhat == nil {
		s.msg.Actor.SendBad(text.TitleFirst(cName), " does not seem to contain ", tName, ".")
		s.msg.Observer.SendInfo("You see ", who, " rummage around in ", cName, ".")
		return
	}

	// Get item's proper name
	tName = attr.FindName(tWhat).Name(tName)

	// Check that the thing doing the taking can carry the item. We do this late
	// in the process so that we have the proper names of the container and the
	// item being taken from it.
	//
	// NOTE: We could just drop the item on the floor if it can't be carried.
	if !tWhere.Found() {
		s.msg.Actor.SendBad("You have nowhere to put ", tName, " if you remove it from ", cName, ".")
		return
	}

	// Check for veto on item being taken
	if veto := attr.FindVetoes(tWhat).Check("TAKE", "GET"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// Check for veto on container
	if veto := attr.FindVetoes(cWhat).Check("TAKE"); veto != nil {
		s.msg.Actor.SendBad(veto.Message())
		return
	}

	// If item is a narrative we can't take it. We do this check after the veto
	// checks as the vetos could give us a better message/reson for not being
	// able to take the item.
	if attr.FindNarrative(tWhat).Found() {
		s.msg.Actor.SendBad("For some reason you cannot take ", tName, " from ", cName, ".")
		s.msg.Observer.SendInfo("You see ", who, " having trouble removing something from ", cName, ".")
		return
	}

	// Move the item from container to our inventory
	if cWhere.Move(tWhat, tWhere) == nil {
		s.msg.Actor.SendBad("Something stops you taking ", tName, " from ", cName, "...")
		s.msg.Observer.SendInfo("You see ", who, " having trouble removing something from ", cName, ".")
		return
	}

	s.msg.Actor.SendGood("You take ", tName, " from ", cName, ".")
	s.msg.Observer.SendInfo("You see ", who, " take something from ", cName, ".")

	s.ok = true
}
