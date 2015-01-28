// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Setup test 'world' with some test data
func Setup() map[string]has.Thing {

	world := map[string]has.Thing{}

	world["cheese"] = Thing(
		NewName("some cheese"),
		NewDescription("This is a blob of very soft, sticky, smelly cheese."),
		NewAlias("cheese"),
		NewVetoes(
			NewVeto("drop", "You can't drop the sticky cheese!"),
		),
	)

	world["mug"] = Thing(
		NewName("a mug"),
		NewDescription("This is a large, white, chipped mug."),
		NewWriting("Stay calm and drink more coffee!"),
		NewAlias("mug"),
		NewAlias("cup"),
		NewInventory(
			Thing(
				NewName("some coffee"),
				NewDescription("This is some hot, strong coffee."),
			),
		),
	)

	world["box"] = Thing(
		NewName("a box"),
		NewDescription("This is a small, wooden box."),
		NewInventory(),
		NewAlias("box"),
	)

	world["bag"] = Thing(
		NewName("a bag"),
		NewDescription("This is a small bag."),
		NewAlias("bag"),
		NewInventory(
			Thing(
				NewName("an apple"),
				NewDescription("This is a juicy red apple."),
				NewAlias("apple"),
			),
			Thing(
				NewName("an orange"),
				NewDescription("This is a large orange."),
				NewAlias("orange"),
			),
		),
	)

	world["plaque"] = Thing(
		NewName("a wooden plaque"),
		NewDescription("This is a small wooden plaque."),
		NewAlias("plaque"),
		NewWriting("Please do not read the plaques!"),
		NewVetoes(
			NewVeto("get", "You can't take the plaque. It's firmly nailed to the wall."),
			NewVeto("examine", "You try to examine the plaque but it makes your head hurt."),
		),
	)

	// Define some locations

	world["loc1"] = Thing(
		NewName("Fireplace"),
		NewDescription("You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. There is a small plaque above the fireplace. To the south the common room extends and east the common room leads to the tavern entrance."),
		NewAlias("tavern", "fireplace"),
		NewInventory(
			world["cheese"],
			world["mug"],
			world["box"],
			world["bag"],
		),
		NewNarrative(
			Thing(
				NewName("an ornate fireplace"),
				NewDescription("This is a very ornate fireplace carved from marble. Either side a dragon curls downward until the head is below the fire looking upward, giving the impression that they are breathing fire."),
				NewAlias("fireplace", "fire"),
			),
			world["plaque"],
		),
		NewExits(),
	)

	world["loc2"] = Thing(
		NewName("Common Room"),
		NewDescription("You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away."),
		NewAlias("tavern", "common"),
		NewInventory(),
		NewExits(),
	)

	world["loc3"] = Thing(
		NewName("Tavern Entrance"),
		NewDescription("You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street."),
		NewAlias("tavern", "entrance"),
		NewInventory(),
		NewExits(),
	)

	world["loc4"] = Thing(
		NewName("Tavern Bar"),
		NewDescription("You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here."),
		NewAlias("tavern", "bar"),
		NewInventory(),
		NewExits(),
	)

	// Link up room exits

	if a := FindExit(world["loc1"]); a != nil {
		a.Link(SOUTH, world["loc2"])
		a.Link(EAST, world["loc3"])
		a.Link(SOUTHEAST, world["loc4"])
	}

	if a := FindExit(world["loc2"]); a != nil {
		a.Link(NORTH, world["loc1"])
		a.Link(NORTHEAST, world["loc3"])
		a.Link(EAST, world["loc4"])
	}

	if a := FindExit(world["loc3"]); a != nil {
		a.Link(WEST, world["loc1"])
		a.Link(SOUTHWEST, world["loc2"])
		a.Link(SOUTH, world["loc4"])
	}

	if a := FindExit(world["loc4"]); a != nil {
		a.Link(NORTH, world["loc3"])
		a.Link(NORTHWEST, world["loc1"])
		a.Link(WEST, world["loc2"])
	}

	return world
}
