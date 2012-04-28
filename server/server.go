package main

import (
	"wolfmud.org/entities"
)

func main() {

	world := entities.NewWorld();

	l1 := entities.NewLocation("Fireplace", "FIREPLACE", "You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. To the south the common room extends and east the common room leads to the tavern entrance.")

	l2 := entities.NewLocation("Common Room", "COMMONROOM", "You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away.")

	l3 := entities.NewLocation("Tavern Bar", "TAVERNBAR", "You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here.")

	l4 := entities.NewLocation("Tavern entrance", "TAVERNENTRANCE", "You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street.")

	// Fireplace
	l1.LinkExit(entities.E, l4)
	l1.LinkExit(entities.SE, l3)
	l1.LinkExit(entities.S, l2)

	// Common room
	l2.LinkExit(entities.N, l1)
	l2.LinkExit(entities.NE, l4)
	l2.LinkExit(entities.E, l3)

	// Tavern Bar
	l3.LinkExit(entities.N, l4)
	l3.LinkExit(entities.W, l2)
	l3.LinkExit(entities.NW, l1)

	// Tavern Entrance
	l4.LinkExit(entities.S, l3)
	l4.LinkExit(entities.SW, l2)
	l4.LinkExit(entities.W, l1)

	world.AddLocation(l1)
	world.AddLocation(l2)
	world.AddLocation(l3)
	world.AddLocation(l4)

	// Some objects
	t1 := entities.NewThing(
		"A curious brass lattice",
		"LATTICE",
		"This is a finely crafted, intricate lattice of fine brass wires forming a roughly ball shaped curiosity.",
	)
	t2 := entities.NewThing(
		"A small ball",
		"BALL",
		"This is a small rubber ball.",
	)

	l1.Add(t1)
	l1.Add(t2)

	world.Start()
}
