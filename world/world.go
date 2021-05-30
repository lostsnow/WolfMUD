// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package world

import (
	"code.wolfmud.org/WolfMUD.git/proc"
)

var World map[string]*proc.Thing

func Load() {

	World = make(map[string]*proc.Thing)
	proc.World = World

	// Items

	cat := proc.NewThing("the tavern cat", "The tavern cat is a ball of fur with one golden eye, the other eye replaced by a large scar. It senses you watching it and returns your gaze with a steady one of its own.")
	cat.Is = proc.NPC
	cat.As[proc.Alias] = "CAT"

	fireplace := proc.NewThing("an ornate fireplace", "This is a very ornate fireplace carved from marble. Either side a dragon curls downward until the head is below the fire looking upward, giving the impression that they are breathing fire.")
	fireplace.Is = proc.Narrative
	fireplace.As[proc.Alias] = "FIREPLACE"

	fire := proc.NewThing("a fire", "Some logs have been placed into the fireplace and are burning away merrily.")
	fire.Is |= proc.Narrative
	fire.As[proc.Alias] = "FIRE"

	greenBall := proc.NewThing("a green ball", "This is a small green ball.")
	greenBall.As[proc.Alias] = "BALL"

	apple := proc.NewThing("an apple", "This is a red apple.")
	apple.As[proc.Alias] = "APPLE"

	bag := proc.NewThing("a bag", "This is a simple cloth bag.")
	bag.Is = proc.Container
	bag.As[proc.Alias] = "BAG"
	bag.In = append(bag.In, apple)

	chest := proc.NewThing("a chest", "This is a large iron bound wooden chest.")
	chest.Is = proc.Container
	chest.As[proc.Alias] = "CHEST"
	chest.In = append(chest.In, greenBall, bag)

	redBall := proc.NewThing("a red ball", "This is a small red ball.")
	redBall.As[proc.Alias] = "BALL"

	note := proc.NewThing("a note", "This is a small piece of paper with something written on it.")
	note.As[proc.Alias] = "NOTE"
	note.As[proc.Writing] = "It says 'Here be dragons'."

	// Locations

	L1 := proc.NewThing("Fireplace", "You are in the corner of the common room in the dragon's breath tavern. A fire burns merrily in an ornate fireplace, giving comfort to weary travellers. The fire causes shadows to flicker and dance around the room, changing darkness to light and back again. To the south the common room continues and east the common room leads to the tavern entrance.")
	L1.Is |= proc.Start
	L1.As[proc.East] = "L3"
	L1.As[proc.Southeast] = "L4"
	L1.As[proc.South] = "L2"
	L1.In = append(L1.In, fireplace, fire, chest)
	World["L1"] = L1

	L2 := proc.NewThing("Common room", "You are in a small, cosy common room in the dragon's breath tavern. Looking around you see a few chairs and tables for patrons. In one corner there is a very old grandfather clock. To the east you see a bar and to the north there is the glow of a fire.")
	L2.As[proc.North] = "L1"
	L2.As[proc.Northeast] = "L3"
	L2.As[proc.East] = "L4"
	L2.In = append(L2.In, cat)
	World["L2"] = L2

	L3 := proc.NewThing("Tavern entrance", "You are in the entryway to the dragon's breath tavern. To the west you see an inviting fireplace and south an even more inviting bar. Eastward a door leads out into the street.")
	L3.As[proc.East] = "L5"
	L3.As[proc.South] = "L4"
	L3.As[proc.Southwest] = "L2"
	L3.As[proc.West] = "L1"
	L3.In = append(L3.In, redBall)
	World["L3"] = L3

	L4 := proc.NewThing("Tavern bar", "You are at the tavern's very sturdy bar. Behind the bar are shelves stacked with many bottles in a dizzying array of sizes, shapes and colours. There are also regular casks of beer, ale, mead, cider and wine behind the bar.")
	L4.As[proc.North] = "L3"
	L4.As[proc.Northwest] = "L1"
	L4.As[proc.West] = "L2"
	L4.In = append(L4.In, note)
	World["L4"] = L4

	L5 := proc.NewThing("Street between tavern and bakers", "You are on a well kept cobbled street. Buildings loom up on either side of you. To the east the smells of a bakery taunt you. To the west the entrance to a tavern. A sign outside the tavern proclaims it to be the \"Dragon's Breath\". The street continues to the north and south.")
	L5.As[proc.North] = "L14"
	L5.As[proc.East] = "L6"
	L5.As[proc.South] = "L7"
	L5.As[proc.West] = "L3"
	World["L5"] = L5
}
