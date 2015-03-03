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

	world["cheese"] = Thing().New(
		Name().New("some cheese"),
		Description().New("This is a blob of very soft, sticky, smelly cheese."),
		Alias().New("cheese"),
		Vetoes().New(
			Veto().New("drop", "You can't drop the sticky cheese!"),
		),
	)

	world["mug"] = Thing().New(
		Name().New("a mug"),
		Description().New("This is a large, white, chipped mug."),
		Writing().New("Stay calm and drink more coffee!"),
		Alias().New("mug"),
		Alias().New("cup"),
		Inventory().New(
			Thing().New(
				Name().New("some coffee"),
				Description().New("This is some hot, strong coffee."),
			),
		),
	)

	world["box"] = Thing().New(
		Name().New("a box"),
		Description().New("This is a small, wooden box."),
		Inventory().New(),
		Alias().New("box"),
	)

	world["bag"] = Thing().New(
		Name().New("a bag"),
		Description().New("This is a small bag."),
		Alias().New("bag"),
		Inventory().New(
			Thing().New(
				Name().New("an apple"),
				Description().New("This is a juicy red apple."),
				Alias().New("apple"),
			),
			Thing().New(
				Name().New("an orange"),
				Description().New("This is a large orange."),
				Alias().New("orange"),
			),
		),
	)

	world["chairs"] = Thing().New(
		Name().New("some rough chairs"),
		Description().New("These chairs(?) are very rough wooden affairs, so rough in fact you decide it's a bad idea to sit on them without some descent rear armour to fend of the splinters."),
		Alias().New("chair", "chairs"),
	)

	world["tables"] = Thing().New(
		Name().New("some rough tables"),
		Description().New("Well you suppose these are tables. If so they must have been made by a blind carpenter having a very bad day."),
		Alias().New("table", "tables"),
	)

	world["plaque"] = Thing().New(
		Name().New("a wooden plaque"),
		Description().New("This is a small wooden plaque."),
		Alias().New("plaque"),
		Writing().New("Please do not read the plaques!"),
		Vetoes().New(
			Veto().New("get", "You can't take the plaque. It's firmly nailed to the wall."),
			Veto().New("examine", "You try to examine the plaque but it makes your head hurt."),
		),
	)

	// Define some locations

	world["loc1"] = Thing().New(
		Name().New("Fireplace"),
		Description().New("You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. There is a small plaque above the fireplace. To the south the common room extends and east the common room leads to the tavern entrance."),
		Alias().New("tavern", "fireplace"),
		Inventory().New(
			world["cheese"],
			world["mug"],
			world["box"],
			world["bag"],
		),
		Narrative().New(
			Thing().New(
				Name().New("an ornate fireplace"),
				Description().New("This is a very ornate fireplace carved from marble. Either side a dragon curls downward until the head is below the fire looking upward, giving the impression that they are breathing fire."),
				Alias().New("fireplace", "fire"),
			),
			world["plaque"],
		),
		Exits().New(),
	)

	world["loc2"] = Thing().New(
		Name().New("Common Room"),
		Description().New("You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away."),
		Alias().New("tavern", "common"),
		Inventory().New(),
		Narrative().New(
			world["chairs"],
			world["tables"],
		),
		Exits().New(),
	)

	world["loc3"] = Thing().New(
		Name().New("Tavern Entrance"),
		Description().New("You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street."),
		Alias().New("tavern", "entrance"),
		Inventory().New(),
		Exits().New(),
	)

	world["loc4"] = Thing().New(
		Name().New("Tavern Bar"),
		Description().New("You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here."),
		Alias().New("tavern", "bar"),
		Inventory().New(),
		Exits().New(),
	)

	// Link up room exits

	if a := Exits().Find(world["loc1"]); a != nil {
		a.Link(SOUTH, world["loc2"])
		a.Link(EAST, world["loc3"])
		a.Link(SOUTHEAST, world["loc4"])
	}

	if a := Exits().Find(world["loc2"]); a != nil {
		a.Link(NORTH, world["loc1"])
		a.Link(NORTHEAST, world["loc3"])
		a.Link(EAST, world["loc4"])
	}

	if a := Exits().Find(world["loc3"]); a != nil {
		a.Link(WEST, world["loc1"])
		a.Link(SOUTHWEST, world["loc2"])
		a.Link(SOUTH, world["loc4"])
	}

	if a := Exits().Find(world["loc4"]); a != nil {
		a.Link(NORTH, world["loc3"])
		a.Link(NORTHWEST, world["loc1"])
		a.Link(WEST, world["loc2"])
	}

	return world
}
