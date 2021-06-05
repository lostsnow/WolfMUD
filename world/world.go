// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package world

import (
	"code.wolfmud.org/WolfMUD.git/proc"
)

// Load creates the game world. This is currently hard-coded for development.
//
// BUG(diddymus): Load will populate proc.World directly as a side effect of
// being called. The proc package can't import the world package as it would
// cause a cyclic import. This should be resolved when we have a proper loader
// written.
func Load() {

	proc.World = make(map[string]*proc.Thing)

	// Items

	cat := proc.NewThing()
	cat.As[proc.Name] = "the tavern cat"
	cat.As[proc.Description] = "The tavern cat is a ball of fur with one golden eye, the other eye replaced by a large scar. It senses you watching it and returns your gaze with a steady one of its own."
	cat.Any[proc.Alias] = []string{"CAT"}
	cat.Any[proc.Veto+"GET"] = []string{"The cat looks at your hand, then looks at you. Hrm, probably a bad idea."}
	cat.Is = proc.NPC

	fireplace := proc.NewThing()
	fireplace.As[proc.Name] = "an ornate fireplace"
	fireplace.As[proc.Description] = "This is a very ornate fireplace carved from marble. Either side a dragon curls downward until the head is below the fire looking upward, giving the impression that they are breathing fire."
	fireplace.Any[proc.Alias] = []string{"FIREPLACE"}
	fireplace.Any[proc.Veto+"GET"] = []string{"For some inexplicable reason you can't just rip out the fireplace and take it!"}
	fireplace.Is = proc.Narrative

	fire := proc.NewThing()
	fire.As[proc.Name] = "a fire"
	fire.As[proc.Description] = "Some logs have been placed into the fireplace and are burning away merrily."
	fire.Any[proc.Alias] = []string{"FIRE"}
	fire.Any[proc.Veto+"GET"] = []string{"Ouch! Hot, hot, hot!"}
	fire.Is |= proc.Narrative

	greenBall := proc.NewThing()
	greenBall.As[proc.Name] = "a green ball"
	greenBall.As[proc.Description] = "This is a small green ball."
	greenBall.Any[proc.Alias] = []string{"+GREEN", "BALL"}

	apple := proc.NewThing()
	apple.As[proc.Name] = "an apple"
	apple.As[proc.Description] = "This is a red apple."
	apple.Any[proc.Alias] = []string{"APPLE"}

	bag := proc.NewThing()
	bag.As[proc.Name] = "a bag"
	bag.As[proc.Description] = "This is a simple cloth bag."
	bag.Any[proc.Alias] = []string{"BAG"}
	bag.Is = proc.Container
	bag.In = append(bag.In, apple)

	chest := proc.NewThing()
	chest.As[proc.Name] = "a chest"
	chest.As[proc.Description] = "This is a large iron bound wooden chest."
	chest.Any[proc.Alias] = []string{"CHEST"}
	chest.Is = proc.Container
	chest.In = append(chest.In, greenBall, bag)

	redBall := proc.NewThing()
	redBall.As[proc.Name] = "a red ball"
	redBall.As[proc.Description] = "This is a small red ball."
	redBall.Any[proc.Alias] = []string{"+RED:BALL"}

	note := proc.NewThing()
	note.As[proc.Name] = "a note"
	note.As[proc.Description] = "This is a small piece of paper with something written on it."
	note.Any[proc.Alias] = []string{"NOTE"}
	note.As[proc.Writing] = "It says 'Here be dragons'."

	door := proc.NewThing()
	door.As[proc.Name] = "the tavern door"
	door.As[proc.Description] = "This is a sturdy wooden door with a simple latch."
	door.Any[proc.Alias] = []string{"+TAVERN", "DOOR"}
	door.As[proc.Blocker] = "E"
	door.As[proc.Where] = "L3"
	door.Is = proc.Narrative

	// Locations

	L1 := proc.NewThing()
	L1.As[proc.Name] = "Fireplace"
	L1.As[proc.Description] = "You are in the corner of the common room in the dragon's breath tavern. A fire burns merrily in an ornate fireplace, giving comfort to weary travellers. The fire causes shadows to flicker and dance around the room, changing darkness to light and back again. To the south the common room continues and east the common room leads to the tavern entrance."
	L1.As[proc.East] = "L3"
	L1.As[proc.Southeast] = "L4"
	L1.As[proc.South] = "L2"
	L1.Is |= proc.Start
	L1.In = append(L1.In, fireplace, fire, chest)
	proc.World["L1"] = L1

	L2 := proc.NewThing()
	L2.As[proc.Name] = "Common room"
	L2.As[proc.Description] = "You are in a small, cosy common room in the dragon's breath tavern. Looking around you see a few chairs and tables for patrons. In one corner there is a very old grandfather clock. To the east you see a bar and to the north there is the glow of a fire."
	L2.As[proc.North] = "L1"
	L2.As[proc.Northeast] = "L3"
	L2.As[proc.East] = "L4"
	L2.In = append(L2.In, cat)
	proc.World["L2"] = L2

	L3 := proc.NewThing()
	L3.As[proc.Name] = "Tavern entrance"
	L3.As[proc.Description] = "You are in the entryway to the dragon's breath tavern. To the west you see an inviting fireplace and south an even more inviting bar. Eastward a door leads out into the street."
	L3.As[proc.East] = "L5"
	L3.As[proc.South] = "L4"
	L3.As[proc.Southwest] = "L2"
	L3.As[proc.West] = "L1"
	L3.In = append(L3.In, redBall, door)
	proc.World["L3"] = L3

	L4 := proc.NewThing()
	L4.As[proc.Name] = "Tavern bar"
	L4.As[proc.Description] = "You are at the tavern's very sturdy bar. Behind the bar are shelves stacked with many bottles in a dizzying array of sizes, shapes and colours. There are also regular casks of beer, ale, mead, cider and wine behind the bar."
	L4.As[proc.North] = "L3"
	L4.As[proc.Northwest] = "L1"
	L4.As[proc.West] = "L2"
	L4.In = append(L4.In, note)
	proc.World["L4"] = L4

	L5 := proc.NewThing()
	L5.As[proc.Name] = "Street between tavern and bakers"
	L5.As[proc.Description] = "You are on a well kept cobbled street. Buildings loom up on either side of you. To the east the smells of a bakery taunt you. To the west the entrance to a tavern. A sign outside the tavern proclaims it to be the \"Dragon's Breath\". The street continues to the north and south."
	L5.As[proc.North] = "L14"
	L5.As[proc.East] = "L6"
	L5.As[proc.South] = "L7"
	L5.As[proc.West] = "L3"
	L5.In = append(L5.In, door)
	proc.World["L5"] = L5
}
