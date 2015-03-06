// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Syntax: PUT item container
func Put(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "You go to put something into something else..."
		return
	}

	var (
		tName = aliases[0]

		tWhat has.Thing
		tInv  has.Inventory
	)

	// Search ourselves for item to put into container
	if tInv = attr.FindInventory(t); tInv != nil {
		tWhat = tInv.Search(tName)
	}

	if tWhat == nil {
		msg = "You have no '" + tName + "' to put into anything."
		return
	}

	// Get item's proper name
	if n := attr.FindName(tWhat); n != nil {
		tName = n.Name()
	}

	// Check a container was specified
	if len(aliases) < 2 {
		msg = "What did you want to put " + tName + " into?"
		return
	}

	// Try and find container
	var (
		cName = aliases[1]

		cWhat  has.Thing
		cWhere has.Thing
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
			if a := attr.FindInventory(cWhere); a != nil {
				cWhat = a.Search(cName)
			}

			// If container not found in inventory also check narratives where we are
			if cWhat == nil {
				if a := attr.FindNarrative(cWhere); a != nil {
					cWhat = a.Search(cName)
				}
			}
		}
	}

	// Was container found?
	if cWhat == nil {
		msg = "You see no '" + cName + "' to put " + tName + " into."
		return
	}

	// Unless our name is Klein we can't put something inside itself! ;)
	if tWhat == cWhat {
		msg = "You can't put " + tName + " inside itself!"
		return
	}

	// Get container's proper name
	if n := attr.FindName(cWhat); n != nil {
		cName = n.Name()
	}

	// Check container is actually a container with an inventory
	cInv := attr.FindInventory(cWhat)
	if cInv == nil {
		msg = "You cannot put " + tName + " into " + cName + "."
		return
	}

	// Check for veto on item being put into container
	if vetoes := attr.Vetoes().Find(tWhat); vetoes != nil {
		if veto := vetoes.Check("DROP", "PUT"); veto != nil {
			msg = veto.Message()
			return
		}
	}

	// Check for veto on container
	if vetoes := attr.Vetoes().Find(cWhat); vetoes != nil {
		if veto := vetoes.Check("PUT"); veto != nil {
			msg = veto.Message()
			return
		}
	}

	// Remove item from where it is
	if tInv.Remove(tWhat) == nil {
		msg = "Something stops you putting " + tName + " anywhere."
		return
	}

	// Put item into comtainer
	cInv.Add(tWhat)

	msg = "You put " + tName + " into " + cName + "."
	return msg, true
}
