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

	l5 := entities.NewLocation("Street between Tavern and Bakers","TAVERNBAKERSSTREET","You are on a well kept cobbled street. Buildings looming up either side of you. To the east the smells of a bakery taunt you, west there is the entrance to a tavern. A sign above the tavern door proclaims it as the Dragon's Breath. The street continues to the north and south.")

	l6 := entities.NewLocation("Baker's Shop","BAKERS","You are standing in a bakers shop. Low tables show an array of fresh breads, cakes and the like. The smells here are beyond description.")

	l7 := entities.NewLocation("Street outside Pawn Shop","PAWNSHOPSTREET","You are on a well kept cobbled street that runs north and south. To the east You can see a small Pawn shop. Southward you can see a large fountain and northward the smell of a bakery teases you.")

	l8 := entities.NewLocation("Pawn Shop","PAWNSHOP","You are in small Pawn shop. All around you on shelves are what looks like a load of useless junk.")

	l9 := entities.NewLocation("Fountain Square","FOUNTAINSQUARE","You are in a small square at the crossing of two roads. In the centre of the square a magnificent fountain has been erected, providing fresh water to any who want it. From here the streets lead off in all directions.")

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
	l4.LinkExit(entities.E, l5)
	l4.LinkExit(entities.S, l3)
	l4.LinkExit(entities.SW, l2)
	l4.LinkExit(entities.W, l1)

	// Street between Tavern and Bakers
	l5.LinkExit(entities.E, l6)
	l5.LinkExit(entities.S, l7)
	l5.LinkExit(entities.W, l4)

	// Bakers
	l6.LinkExit(entities.W, l5)

	// Street outside Pawn Shop
	l7.LinkExit(entities.N, l5)
	l7.LinkExit(entities.E, l8)
	l7.LinkExit(entities.S, l9)

	// Pawn Shop
	l8.LinkExit(entities.W, l7)

	// Fountain Square
	l9.LinkExit(entities.N, l7)
// ???

	world.AddLocation(l1)
	world.AddLocation(l2)
	world.AddLocation(l3)
	world.AddLocation(l4)
	world.AddLocation(l5)
	world.AddLocation(l6)
	world.AddLocation(l7)
	world.AddLocation(l8)
	world.AddLocation(l9)

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
