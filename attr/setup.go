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

	world["cheese"] = NewThing(
		NewName("some cheese"),
		NewDescription("This is a blob of very soft, sticky, smelly cheese."),
		NewAlias("cheese"),
		NewVetoes(
			NewVeto("drop", "You can't drop the sticky cheese!"),
		),
	)

	world["mug"] = NewThing(
		NewName("a mug"),
		NewDescription("This is a large, white, chipped mug."),
		NewWriting("Stay calm and drink more coffee!"),
		NewAlias("mug"),
		NewAlias("cup"),
		NewInventory(
			NewThing(
				NewName("some coffee"),
				NewDescription("This is some hot, strong coffee."),
			),
		),
	)

	world["box"] = NewThing(
		NewName("a box"),
		NewDescription("This is a small, wooden box."),
		NewInventory(),
		NewAlias("box"),
	)

	world["bag"] = NewThing(
		NewName("a bag"),
		NewDescription("This is a small bag."),
		NewAlias("bag"),
		NewInventory(
			NewThing(
				NewName("an apple"),
				NewDescription("This is a juicy red apple."),
				NewAlias("apple"),
			),
			NewThing(
				NewName("an orange"),
				NewDescription("This is a large orange."),
				NewAlias("orange"),
			),
		),
	)

	world["chairs"] = NewThing(
		NewName("some rough chairs"),
		NewDescription("These chairs(?) are very rough wooden affairs, so rough in fact you decide it's a bad idea to sit on them without some descent rear armour to fend of the splinters."),
		NewAlias("chair", "chairs"),
	)

	world["tables"] = NewThing(
		NewName("some rough tables"),
		NewDescription("Well you suppose these are tables. If so they must have been made by a blind carpenter having a very bad day."),
		NewAlias("table", "tables"),
	)

	world["plaque"] = NewThing(
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

	world["loc1"] = NewThing(
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
			NewThing(
				NewName("an ornate fireplace"),
				NewDescription("This is a very ornate fireplace carved from marble. Either side a dragon curls downward until the head is below the fire looking upward, giving the impression that they are breathing fire."),
				NewAlias("fireplace", "fire"),
			),
			world["plaque"],
		),
		NewExits(),
	)

	world["loc2"] = NewThing(
		NewName("Common Room"),
		NewDescription("You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away."),
		NewAlias("tavern", "common"),
		NewInventory(),
		NewNarrative(
			world["chairs"],
			world["tables"],
		),
		NewExits(),
	)

	world["loc3"] = NewThing(
		NewName("Tavern Entrance"),
		NewDescription("You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street."),
		NewAlias("tavern", "entrance"),
		NewInventory(),
		NewExits(),
	)

	world["loc4"] = NewThing(
		NewName("Tavern Bar"),
		NewDescription("You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here."),
		NewAlias("tavern", "bar"),
		NewInventory(),
		NewExits(),
	)

	world["loc5"] = NewThing(
		NewName("Street between Tavern and Bakers"),
		NewDescription("You are on a well kept cobbled street. Buildings looming up either side of you. To the east the smells of a bakery taunt you, west there is the entrance to a tavern. A sign above the tavern door proclaims it as the Dragon's Breath. The street continues to the north and south."),
		NewAlias("tavern", "bakers", "street"),
		NewInventory(),
		NewExits(),
	)

	// Link up room exits

	if a := FindExits(world["loc1"]); a != nil {
		a.Link(SOUTH, world["loc2"])
		a.Link(EAST, world["loc3"])
		a.Link(SOUTHEAST, world["loc4"])
	}

	if a := FindExits(world["loc2"]); a != nil {
		a.Link(NORTH, world["loc1"])
		a.Link(NORTHEAST, world["loc3"])
		a.Link(EAST, world["loc4"])
	}

	if a := FindExits(world["loc3"]); a != nil {
		a.Link(WEST, world["loc1"])
		a.Link(SOUTHWEST, world["loc2"])
		a.Link(SOUTH, world["loc4"])
		a.Link(EAST, world["loc5"])
	}

	if a := FindExits(world["loc4"]); a != nil {
		a.Link(NORTH, world["loc3"])
		a.Link(NORTHWEST, world["loc1"])
		a.Link(WEST, world["loc2"])
	}

	if a := FindExits(world["loc5"]); a != nil {
		a.Link(WEST, world["loc3"])
	}

	return world
}
