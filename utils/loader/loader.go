// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package loader loads entities into the given world.
//
// TODO: The loader should read text files and parse them creating entities
// that are then loaded into the world. At the moment the file parser has not
// been written and the loader is hardcoded.
package loader

import (
	"wolfmud.org/entities/location"
	"wolfmud.org/entities/thing"
	"wolfmud.org/entities/world"
)

// Load creates entities and adds them to the given world.
func Load(world *world.World) {

	l1 := location.New("Fireplace", []string{"FIREPLACE"}, "You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. To the south the common room extends and east the common room leads to the tavern entrance.")

	l2 := location.New("Common Room", []string{"COMMONROOM"}, "You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away.")

	l3 := location.New("Tavern Bar", []string{"TAVERNBAR"}, "You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here.")

	l4 := location.New("Tavern entrance", []string{"TAVERNENTRANCE"}, "You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street.")

	l5 := location.New("Street between Tavern and Bakers", []string{"TAVERNBAKERSSTREET"}, "You are on a well kept cobbled street. Buildings looming up either side of you. To the east the smells of a bakery taunt you, west there is the entrance to a tavern. A sign above the tavern door proclaims it as the Dragon's Breath. The street continues to the north and south.")

	l6 := location.New("Baker's Shop", []string{"BAKERS"}, "You are standing in a bakers shop. Low tables show an array of fresh breads, cakes and the like. The smells here are beyond description.")

	l7 := location.New("Street outside Pawn Shop", []string{"PAWNSHOPSTREET"}, "You are on a well kept cobbled street that runs north and south. To the east You can see a small Pawn shop. Southward you can see a large fountain and northward the smell of a bakery teases you.")

	l8 := location.New("Pawn Shop", []string{"PAWNSHOP"}, "You are in small Pawn shop. All around you on shelves are what looks like a load of useless junk.")

	l9 := location.New("Fountain Square", []string{"FOUNTAINSQUARE"}, "You are in a small square at the crossing of two roads. In the centre of the square a magnificent fountain has been erected, providing fresh water to any who want it. From here the streets lead off in all directions.")

	// Fireplace
	l1.LinkExit(location.E, l4)
	l1.LinkExit(location.SE, l3)
	l1.LinkExit(location.S, l2)

	// Common room
	l2.LinkExit(location.N, l1)
	l2.LinkExit(location.NE, l4)
	l2.LinkExit(location.E, l3)

	// Tavern Bar
	l3.LinkExit(location.N, l4)
	l3.LinkExit(location.W, l2)
	l3.LinkExit(location.NW, l1)

	// Tavern Entrance
	l4.LinkExit(location.E, l5)
	l4.LinkExit(location.S, l3)
	l4.LinkExit(location.SW, l2)
	l4.LinkExit(location.W, l1)

	// Street between Tavern and Bakers
	l5.LinkExit(location.E, l6)
	l5.LinkExit(location.S, l7)
	l5.LinkExit(location.W, l4)

	// Bakers
	l6.LinkExit(location.W, l5)

	// Street outside Pawn Shop
	l7.LinkExit(location.N, l5)
	l7.LinkExit(location.E, l8)
	l7.LinkExit(location.S, l9)

	// Pawn Shop
	l8.LinkExit(location.W, l7)

	// Fountain Square
	l9.LinkExit(location.N, l7)
	// ???
	//
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
	t1 := thing.New(
		"A curious brass lattice",
		[]string{"LATTICE"},
		"This is a finely crafted, intricate lattice of fine brass wires forming a roughly ball shaped curiosity.",
	)
	t2 := thing.New(
		"A small ball",
		[]string{"BALL"},
		"This is a small rubber ball.",
	)

	l1.Add(t1)
	l1.Add(t2)
}
