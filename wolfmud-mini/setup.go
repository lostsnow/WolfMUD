// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package main

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

// Setup test 'world' with some test data
func setup() map[string]has.Thing {

	world := map[string]has.Thing{}

	world["cheese"] = attr.Thing(
		attr.NewName("some cheese"),
		attr.NewDescription("This is a blob of very soft, sticky, smelly cheese."),
		attr.NewAlias("cheese"),
		attr.NewVeto(
			[][2]string{
				{"drop", "You can't drop the sticky cheese!"},
			},
		),
	)

	world["mug"] = attr.Thing(
		attr.NewName("a mug"),
		attr.NewDescription("This is a large mug. It has some writing on it."),
		attr.NewWriting("Stay calm and drink more coffee!"),
		attr.NewAlias("mug"),
		attr.NewInventory(
			attr.Thing(
				attr.NewName("some coffee"),
				attr.NewDescription("This is some hot, strong coffee."),
			),
		),
	)

	world["box"] = attr.Thing(
		attr.NewName("a box"),
		attr.NewDescription("This is a small, wooden box."),
		attr.NewInventory(),
		attr.NewAlias("box"),
	)

	world["bag"] = attr.Thing(
		attr.NewName("a bag"),
		attr.NewDescription("This is a small bag."),
		attr.NewAlias("bag"),
		attr.NewInventory(
			attr.Thing(
				attr.NewName("an apple"),
				attr.NewDescription("This is a juicy red apple."),
				attr.NewAlias("apple"),
			),
			attr.Thing(
				attr.NewName("an orange"),
				attr.NewDescription("This is a large orange."),
				attr.NewAlias("orange"),
			),
		),
	)

	world["plaque"] = attr.Thing(
		attr.NewName("a wooden plaque"),
		attr.NewDescription("This is a small wooden plaque with some writing on it."),
		attr.NewAlias("plaque"),
		attr.NewWriting("Please do not read the plaques!"),
		attr.NewVeto(
			[][2]string{
				{"get", "You can't take the plaque. It's firmly nailed to the wall."},
				{"examine", "You try to examine the plaque but it makes your head hurt."},
			},
		),
	)

	// Define some locations

	world["loc1"] = attr.Thing(
		attr.NewName("Fireplace"),
		attr.NewDescription("You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. There is a small plaque above the fireplace. To the south the common room extends and east the common room leads to the tavern entrance."),
		attr.NewAlias("tavern", "fireplace"),
		attr.NewInventory(
			world["cheese"],
			world["mug"],
			world["box"],
			world["bag"],
		),
		attr.NewNarrative(
			attr.Thing(
				attr.NewName("an ornate fireplace"),
				attr.NewDescription("This is a very ornate fireplace carved from marble. Either side a dragon curls downward until the head is below the fire looking upward, giving the impression that they are breathing fire."),
				attr.NewAlias("fireplace", "fire"),
			),
			world["plaque"],
		),
		attr.NewExits(),
	)

	world["loc2"] = attr.Thing(
		attr.NewName("Common Room"),
		attr.NewDescription("You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away."),
		attr.NewAlias("tavern", "common"),
		attr.NewInventory(),
		attr.NewExits(),
	)

	world["loc3"] = attr.Thing(
		attr.NewName("Tavern Entrance"),
		attr.NewDescription("You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street."),
		attr.NewAlias("tavern", "entrance"),
		attr.NewInventory(),
		attr.NewExits(),
	)

	world["loc4"] = attr.Thing(
		attr.NewName("Tavern Bar"),
		attr.NewDescription("You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here."),
		attr.NewAlias("tavern", "bar"),
		attr.NewInventory(),
		attr.NewExits(),
	)

	// Link up room exits

	if a := attr.FindExit(world["loc1"]); a != nil {
		a.Link(attr.SOUTH, world["loc2"])
		a.Link(attr.EAST, world["loc3"])
		a.Link(attr.SOUTHEAST, world["loc4"])
	}

	if a := attr.FindExit(world["loc2"]); a != nil {
		a.Link(attr.NORTH, world["loc1"])
		a.Link(attr.NORTHEAST, world["loc3"])
		a.Link(attr.EAST, world["loc4"])
	}

	if a := attr.FindExit(world["loc3"]); a != nil {
		a.Link(attr.WEST, world["loc1"])
		a.Link(attr.SOUTHWEST, world["loc2"])
		a.Link(attr.SOUTH, world["loc4"])
	}

	if a := attr.FindExit(world["loc4"]); a != nil {
		a.Link(attr.NORTH, world["loc3"])
		a.Link(attr.NORTHWEST, world["loc1"])
		a.Link(attr.WEST, world["loc2"])
	}

	return world
}
