// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Syntax: TAKE item container
func init() {
	AddHandler(Take, "TAKE")
}

func Take(s *state) {

	if len(s.words) == 0 {
		s.msg.actor.WriteString("You go to take something out of something else...")
		return
	}

	tName := s.words[0]

	// Was container specified? We have to check for the container first as the
	// item would be in the container, if there is no container specified we
	// cannot find the item and hence resolve the proper name for it.
	if len(s.words) < 2 {
		s.msg.actor.WriteJoin("What did you want to take '", tName, "' out of?")
		return
	}

	var (
		cName = s.words[1]

		cWhat has.Thing
	)

	// Find the taking things own inventory. We remember this inventory as this
	// is where the item will be put if sucessfully taken from the container
	tWhere := attr.FindInventory(s.actor)

	// If we found and inventory search it for the container
	if tWhere != nil {
		cWhat = tWhere.Search(cName)
	}

	// If container not found yet work out where we are and search there
	if cWhat == nil {
		var where has.Inventory

		if a := attr.FindLocate(s.actor); a != nil {
			where = a.Where()
		}

		// If we are nowhere we are not going to find the container so bail early
		if where == nil {
			s.msg.actor.WriteJoin("You see no '", cName, "' to take anything from.")
			return
		}

		// Search for container in the inventory where we are
		cWhat = where.Search(cName)

		// If container not found in inventory also check narratives where we are
		if cWhat == nil {
			if a := attr.FindNarrative(where.Parent()); a != nil {
				cWhat = a.Search(cName)
			}
		}
	}

	// Was container found?
	if cWhat == nil {
		s.msg.actor.WriteJoin("You see no '", cName, "' to take things out of.")
		return
	}

	// Get container's proper name
	if n := attr.FindName(cWhat); n != nil {
		cName = n.Name()
	}

	// Check container is actually a container with an inventory
	cInv := attr.FindInventory(cWhat)
	if cInv == nil {
		s.msg.actor.WriteJoin("You cannot take anything from ", cName)
		return
	}

	// Is item to be taken in the container?
	tWhat := cInv.Search(tName)
	if tWhat == nil {
		s.msg.actor.WriteJoin("There is no '", tName, "' in ", cName, ".")
		return
	}

	// Get item's proper name
	if n := attr.FindName(tWhat); n != nil {
		tName = n.Name()
	}

	// Check that the thing doing the taking can carry the item. We do this late
	// in the process so that we have the proper names of the container and the
	// item being taken from it.
	//
	// NOTE: We could just drop the item on the floor if it can't be carried.
	if tWhere == nil {
		s.msg.actor.WriteJoin("You have nowhere to put ", tName, " if you remove it from ", cName, ".")
		return
	}

	// Check for veto on item being taken
	if vetoes := attr.FindVetoes(tWhat); vetoes != nil {
		if veto := vetoes.Check("TAKE", "GET"); veto != nil {
			s.msg.actor.WriteString(veto.Message())
			return
		}
	}

	// Check for veto on container
	if vetoes := attr.FindVetoes(cWhat); vetoes != nil {
		if veto := vetoes.Check("TAKE"); veto != nil {
			s.msg.actor.WriteString(veto.Message())
			return
		}
	}

	// Try and remove the item from container
	if cInv.Remove(tWhat) == nil {
		s.msg.actor.WriteJoin("Something stops you taking ", tName, " from ", cName, "...")
		return
	}

	// Put item in taking thing's inventory
	tWhere.Add(tWhat)

	s.msg.actor.WriteJoin("You take ", tName, " from ", cName, ".")
	s.ok = true
}
