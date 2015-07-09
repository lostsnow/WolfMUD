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
func Take(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "You go to take something out of something else..."
		return
	}

	tName := aliases[0]

	// Was container specified? (Item would be in container, cannot resolve tName)
	if len(aliases) < 2 {
		msg = "What did you want to take '" + tName + "' out of?"
		return
	}

	var (
		cName = aliases[1]

		cWhat  has.Thing
		cWhere has.Inventory
	)

	// Search ourselves for container to get something from
	from := attr.FindInventory(t)
	if from != nil {
		cWhat = from.Search(cName)
	}

	// Container not found?
	if cWhat == nil {

		// Work out where we are
		if a := attr.FindLocate(t); a != nil {
			cWhere = a.Where()
		}

		// If we are somewhere then check around us
		if cWhere != nil {

			// Search for container in the inventory where we are
			cWhat = cWhere.Search(cName)

			// If container not found in inventory also check narratives where we are
			if cWhat == nil {
				if a := attr.FindNarrative(cWhere.Parent()); a != nil {
					cWhat = a.Search(cName)
				}
			}
		}
	}

	// Was container found?
	if cWhat == nil {
		msg = "You see no '" + cName + "' to take things out of."
		return
	}

	// Get container's proper name
	if n := attr.FindName(cWhat); n != nil {
		cName = n.Name()
	}

	// Check container is actually a container with an inventory
	cInv := attr.FindInventory(cWhat)
	if cInv == nil {
		msg = "You cannot take anything from " + cName
		return
	}

	// Is item to be taken in the container?
	tWhat := cInv.Search(tName)
	if tWhat == nil {
		msg = "There is no '" + tName + "' in " + cName + "."
		return
	}

	// Get item's proper name
	if n := attr.FindName(tWhat); n != nil {
		tName = n.Name()
	}

	// Check for veto on item being taken
	if vetoes := attr.FindVetoes(tWhat); vetoes != nil {
		if veto := vetoes.Check("TAKE"); veto != nil {
			msg = veto.Message()
			return
		}
	}

	// Check for veto on container
	if vetoes := attr.FindVetoes(cWhat); vetoes != nil {
		if veto := vetoes.Check("TAKE"); veto != nil {
			msg = veto.Message()
			return
		}
	}

	// Find inventory of thing doing the taking
	// NOTE: We could drop the item on the floor if it can't be carried.
	tInv := attr.FindInventory(t)
	if tInv == nil {
		msg = "You have nowhere to put " + tName + " if you remove it from " + cName + "."
		return
	}

	// Remove item from container
	if cInv.Remove(tWhat) == nil {
		msg = "Something stops you taking " + tName + " from " + cName + "..."
		return
	}

	// Put item in taking thing's inventory
	tInv.Add(tWhat)

	msg = "You take " + tName + " from " + cName + "."
	return msg, true
}
