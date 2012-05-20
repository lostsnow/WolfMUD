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

	l1 := location.New("Fireplace", []string{"TAVERN", "FIREPLACE"}, "You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. To the south the common room extends and east the common room leads to the tavern entrance.")

	l2 := location.New("Common Room", []string{"TAVERN", "COMMON"}, "You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away.")

	l3 := location.New("Tavern entrance", []string{"TAVERN", "ENTRANCE"}, "You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street.")

	l4 := location.New("Tavern Bar", []string{"TAVERN", "BAR"}, "You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here.")

	l5 := location.New("Street between Tavern and Bakers", []string{"TAVERN", "BAKERS", "STREET"}, "You are on a well kept cobbled street. Buildings looming up either side of you. To the east the smells of a bakery taunt you, west there is the entrance to a tavern. A sign above the tavern door proclaims it as the Dragon's Breath. The street continues to the north and south.")

	l6 := location.New("Baker's Shop", []string{"BAKERS"}, "You are standing in a bakers shop. Low tables show an array of fresh breads, cakes and the like. The smells here are beyond description.")

	l7 := location.New("Street outside Pawn Shop", []string{"PAWNSHOPSTREET"}, "You are on a well kept cobbled street that runs north and south. To the east You can see a small Pawn shop. Southward you can see a large fountain and northward the smell of a bakery teases you.")

	l8 := location.New("Pawn Shop", []string{"PAWNSHOP"}, "You are in small Pawn shop. All around you on shelves are what looks like a load of useless junk.")

	l9 := location.New("Fountain Square", []string{"FOUNTAIN", "SQUARE"}, "You are in a small square at the crossing of two roads. In the centre of the square a magnificent fountain has been erected, providing fresh water to any who want it. From here the streets lead off in all directions.")

	l10 := location.New("Street outside Armourer", []string{"STREET", "ARMOURER"}, "You are on a well kept cobbled street which runs to the east and west. To the south you can see the shop of an armourer.")

	l11 := location.New("Armourer's", []string{"ARMOURER"}, "You are in a small Armourers. Here, if you have the money, you could stagger out weighed down by more armour than a tank.")

	l12 := location.New("Street outside Weapon Shop", []string{"WEAPONSHOP", "STREET"}, "You are on a well kept, wide street which runs to the east and west. To the south you can see a weapon shop.")

	l13 := location.New("Weapons Shop", []string{"WEAPONS", "SHOP"}, "You are in a small weapons shop. If it's 'big gun' stuff you're after you would do better looking else where.")

	l14 := location.New("Crossroads", []string{"CROSSROADS"}, "You are at the cross roads of two streets. One street runs east to west and the other north to south.")

	l15 := location.New("Street outside Trading Post", []string{"STREET", "TRADINGPOST"}, "You are on a street running east to west. To north is a large Trading Post.")

	// Fireplace
	l1.LinkExit(location.E, l3)
	l1.LinkExit(location.SE, l4)
	l1.LinkExit(location.S, l2)

	// Common room
	l2.LinkExit(location.N, l1)
	l2.LinkExit(location.NE, l3)
	l2.LinkExit(location.E, l4)

	// Tavern Entrance
	l3.LinkExit(location.E, l5)
	l3.LinkExit(location.S, l4)
	l3.LinkExit(location.SW, l2)
	l3.LinkExit(location.W, l1)

	// Tavern Bar
	l4.LinkExit(location.N, l3)
	l4.LinkExit(location.NW, l1)
	l4.LinkExit(location.W, l2)

	// Street between Tavern and Bakers
	l5.LinkExit(location.E, l6)
	l5.LinkExit(location.S, l7)
	l5.LinkExit(location.W, l3)

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
	l9.LinkExit(location.E, l12)
	l9.LinkExit(location.W, l10)

	// Street outside Armourer
	l10.LinkExit(location.E, l9)
	l10.LinkExit(location.S, l11)
	//W24

	// Armourer's
	l11.LinkExit(location.N, l10)

	// Street outside Weapon Shop
	l12.LinkExit(location.S, l13)
	l12.LinkExit(location.W, l9)
	//E21

	// Weapons Shop
	l13.LinkExit(location.N, l12)

	// Crossroads
	l14.LinkExit(location.E, l15)
	l14.LinkExit(location.S, l5)
	// S17
	// N45

	// Street outside Trading Post
	l15.LinkExit(location.W, l14)
	// N16
	// E19

	world.AddLocation(l1)
	world.AddLocation(l2)
	world.AddLocation(l3)
	world.AddLocation(l4)
	world.AddLocation(l5)
	world.AddLocation(l6)
	world.AddLocation(l7)
	world.AddLocation(l8)
	world.AddLocation(l9)
	world.AddLocation(l10)
	world.AddLocation(l11)
	world.AddLocation(l12)
	world.AddLocation(l13)
	world.AddLocation(l14)
	world.AddLocation(l15)

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
