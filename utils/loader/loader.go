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
	"wolfmud.org/entities/thing/item"
	"wolfmud.org/entities/world"
)

// Load creates entities and adds them to the given world.
func Load(world *world.World) {

	l1 := location.NewStart("Fireplace", []string{"TAVERN", "FIREPLACE"}, "You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. To the south the common room extends and east the common room leads to the tavern entrance.")

	l2 := location.NewBasic("Common Room", []string{"TAVERN", "COMMON"}, "You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away.")

	l3 := location.NewBasic("Tavern entrance", []string{"TAVERN", "ENTRANCE"}, "You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street.")

	l4 := location.NewBasic("Tavern Bar", []string{"TAVERN", "BAR"}, "You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here.")

	l5 := location.NewBasic("Street between Tavern and Bakers", []string{"TAVERN", "BAKERS", "STREET"}, "You are on a well kept cobbled street. Buildings looming up either side of you. To the east the smells of a bakery taunt you, west there is the entrance to a tavern. A sign above the tavern door proclaims it as the Dragon's Breath. The street continues to the north and south.")

	l6 := location.NewBasic("Baker's Shop", []string{"BAKERS"}, "You are standing in a bakers shop. Low tables show an array of fresh breads, cakes and the like. The smells here are beyond description.")

	l7 := location.NewBasic("Street outside Pawn Shop", []string{"PAWNSHOPSTREET"}, "You are on a well kept cobbled street that runs north and south. To the east You can see a small Pawn shop. Southward you can see a large fountain and northward the smell of a bakery teases you.")

	l8 := location.NewBasic("Pawn Shop", []string{"PAWNSHOP"}, "You are in small Pawn shop. All around you on shelves are what looks like a load of useless junk.")

	l9 := location.NewBasic("Fountain Square", []string{"FOUNTAIN", "SQUARE"}, "You are in a small square at the crossing of two roads. In the centre of the square a magnificent fountain has been erected, providing fresh water to any who want it. From here the streets lead off in all directions.")

	l10 := location.NewBasic("Street outside Armourer", []string{"STREET", "ARMOURER"}, "You are on a well kept cobbled street which runs to the east and west. To the south you can see the shop of an armourer.")

	l11 := location.NewBasic("Armourer's", []string{"ARMOURER"}, "You are in a small Armourers. Here, if you have the money, you could stagger out weighed down by more armour than a tank.")

	l12 := location.NewBasic("Street outside Weapon Shop", []string{"WEAPONSHOP", "STREET"}, "You are on a well kept, wide street which runs to the east and west. To the south you can see a weapon shop.")

	l13 := location.NewBasic("Weapons Shop", []string{"WEAPONS", "SHOP"}, "You are in a small weapons shop. If it's 'big gun' stuff you're after you would do better looking else where.")

	l14 := location.NewBasic("Crossroads", []string{"CROSSROADS"}, "You are at the cross roads of two streets. One street runs east to west and the other north to south.")

	l15 := location.NewBasic("Street outside Trading Post", []string{"STREET", "TRADINGPOST"}, "You are on a street running east to west. To north is a large Trading Post.")

	l16 := location.NewBasic("Trading Post", []string{"TRADINGPOST", "SHOP"}, "You are standing in a large Trading Post . The only exit is west into the street.")

	l17 := location.NewBasic("Street outside Guard House", []string{"STREET", "GUARDHOUSE"}, "You are on a street running east to west. To north is the cities Guard House.")

	l18 := location.NewBasic("Guard House", []string{"GUARDHOUSE", "STREET"}, "You are in the cities Guard House.")

	l19 := location.NewBasic("Street by North Bridge", []string{"STREET"}, "You are at a junction in the street. You can either head south, east or west. East there is the north bridge over the cities river.")

	l20 := location.NewStart("North Bridge", []string{"NORTHBRIDGE"}, "You are standing on the west side of an incomplete bridge. By the looks of it the city wants to expand onto the far banks of the river. Down river to the south you can see another bridge in a similar state of construction.")

	l21 := location.NewBasic("Street by South Bridge", []string{"STREET"}, "You are at a junction in the street. You can either head north, east or west. East there is the south bridge over the cities river.")

	l22 := location.NewStart("South Bridge", []string{"SOUTHBRIDGE"}, "You are standing on the west side of an incomplete bridge. By the looks of it the city wants to expand onto the far banks of the river. Up river to the north you can see another bridge in a similar state of construction.")

	l23 := location.NewBasic("Money Changer's Office", []string{"MONEYCHANGER", "SHOP"}, "You are standing in the small office of a Money Changer. The only exit is north into the street.")

	l24 := location.NewBasic("Street outside Money Changer", []string{"STREET"}, "You are in the street outside a small Money Changer's Office. The street heads to the east and west. The entrance to the Money Changer is to the South.")

	l25 := location.NewBasic("Street", []string{"STREET"}, "You are in the street which is running east and west. To the north looms a dark alley.")

	l26 := location.NewBasic("Street outside City Gardens", []string{"STREET"}, "You are at the end of street which runs to the east. To the west you can see the entrance to the City Gardens. A narrow alley runs south.")

	l27 := location.NewBasic("Entrance to City Gardens", []string{"GARDENS", "ENTRANCE"}, "You are standing at the entrance to the City Gardens. Around you flowers blossom and insects hum. West you can see an intricate iron gate leading into the gardens. East the streets await.")

	l28 := location.NewBasic("Gardens", []string{"GARDENS"}, "You are in a fine, formal garden. Around you small trees and shrubs grow. Here and there a splash of colour is present in the form of small delicate flowers. To the east you can see the gateway, northward there is a fish pond.")

	l29 := location.NewBasic("Quiet Sheltered Area", []string{"GARDENS", "SHELTERED"}, "You are in a small quiet area of the garden. Tall bushes have grown up over a trellis work to provide a small shaded area to come and sit in. The only exit from this area is southward.")

	l30 := location.NewBasic("Gardens", []string{"GARDENS"}, "You are in a fine, formal garden. Around you small trees and shrubs grow. Here and there a splash of colour is present in the form of small delicate flowers. The gardens continue north, south and west.")

	l31 := location.NewBasic("Fishpond", []string{"GARDENS", "FISHPOND"}, "You are standing next to a small fishpond. Paths lead off north, south and west deeper into the gardens.")

	l32 := location.NewBasic("Gardens", []string{"GARDENS"}, "You are in a fine, formal garden. Around you small trees and shrubs grow. Here and there a splash of colour is present in the form of small delicate flowers. The gardens continue north, south and east.")

	l33 := location.NewBasic("Gardens", []string{"GARDENS"}, "You are in a fine, formal garden. Around you small trees and shrubs grow. Here and there a splash of colour is present in the form of small delicate flowers. The gardens continue east, west and south.")

	l34 := location.NewBasic("Quiet Sheltered Area", []string{"GARDENS", "SHELTERED"}, "You are in a small quiet area of the garden. Tall bushes have grown up over a trellis work to provide a small shaded area to come and sit in. The only exit from this area is eastward.")

	l35 := location.NewBasic("Gravel Path", []string{"GARDENS", "PATH"}, "You find yourself on a narrow gravel path leading between some bushes. The path continues south or you can go north into the gardens.")

	l36 := location.NewBasic("Secluded Path by Shed", []string{"GARDENS", "PATH", "SECLUDED"}, "You are on a small secluded gravel path screened off from the formal gardens by some large bushes. To the south you can make out the door to a small shed. A rock has been positioned by the shed, from the looks of it by someone who thought it might be artistic. The gravel path you are on leads northward.")

	l37 := location.NewBasic("Garden Shed", []string{"GARDENS", "SHED"}, "You are in a small garden shed. The only exit appears to be through the door to the north.")

	l38 := location.NewBasic("Dim Alley", []string{"ALLEY"}, "You are in a dim alley full of rubbish. The alley continues northward. To the south it leads into the street.")

	l39 := location.NewBasic("Dim Alley", []string{"ALLEY"}, "You are in a dim alley full of rubbish. The alley continues north and south.")

	l40 := location.NewBasic("Bend in Dim Alley", []string{"ALLEY"}, "You are in a dim alley full of rubbish. Here the alley bends to the south and to the west.")

	l41 := location.NewBasic("Rogue's Den Entrance", []string{"DEN", "ENTRANCE"}, "You are in a dim alley which leads to the east. Partially hidden by rubbish you can see a small trap door.")

	l42 := location.NewBasic("Rogue's Den", []string{"DEN"}, "You are in a large dim room. Looking around to catch glimpses of things moving in the shadows.")

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
	l5.LinkExit(location.N, l14)
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
	l10.LinkExit(location.W, l24)

	// Armourer's
	l11.LinkExit(location.N, l10)

	// Street outside Weapon Shop
	l12.LinkExit(location.S, l13)
	l12.LinkExit(location.W, l9)
	l12.LinkExit(location.E, l21)

	// Weapons Shop
	l13.LinkExit(location.N, l12)

	// Crossroads
	l14.LinkExit(location.E, l15)
	l14.LinkExit(location.S, l5)
	l14.LinkExit(location.W, l17)
	// N45

	// Street outside Trading Post
	l15.LinkExit(location.W, l14)
	l15.LinkExit(location.N, l16)
	l15.LinkExit(location.E, l19)

	// Trading Post
	l16.LinkExit(location.S, l15)

	// Street outside Guard House
	l17.LinkExit(location.N, l18)
	l17.LinkExit(location.E, l14)
	l17.LinkExit(location.W, l25)

	// Guard House
	l18.LinkExit(location.S, l17)

	// Street by North Bridge
	l19.LinkExit(location.E, l20)
	l19.LinkExit(location.W, l15)
	// S43

	// North Bridge
	l20.LinkExit(location.W, l19)

	// Street by South Bridge
	l21.LinkExit(location.E, l22)
	l21.LinkExit(location.W, l12)
	// N44

	// South Bridge
	l22.LinkExit(location.W, l21)

	// Money Changer's Office
	l23.LinkExit(location.N, l24)

	// Street outside Money Changer
	l24.LinkExit(location.E, l10)
	l24.LinkExit(location.S, l23)
	// W60

	// Street
	l25.LinkExit(location.E, l17)
	l25.LinkExit(location.W, l26)
	l25.LinkExit(location.N, l38)

	// Street outside City Gardens
	l26.LinkExit(location.E, l25)
	l26.LinkExit(location.W, l27)
	// S58

	// Entrance to City Gardens
	l27.LinkExit(location.E, l26)
	l27.LinkExit(location.W, l28)

	// Gardens
	l28.LinkExit(location.E, l27)
	l28.LinkExit(location.N, l31)

	// Quiet Sheltered Area
	l29.LinkExit(location.S, l30)

	// Gardens
	l30.LinkExit(location.N, l29)
	l30.LinkExit(location.S, l31)
	l30.LinkExit(location.W, l33)

	// Fishpond
	l31.LinkExit(location.N, l30)
	l31.LinkExit(location.S, l28)
	l31.LinkExit(location.W, l32)

	// Gardens
	l32.LinkExit(location.N, l33)
	l32.LinkExit(location.E, l31)
	l32.LinkExit(location.S, l35)

	// Gardens
	l33.LinkExit(location.E, l30)
	l33.LinkExit(location.S, l32)
	l33.LinkExit(location.W, l34)

	// Quiet Sheltered Area
	l34.LinkExit(location.E, l33)

	// Gravel Path
	l35.LinkExit(location.N, l32)
	l35.LinkExit(location.S, l36)

	// Secluded Path by Shed
	l36.LinkExit(location.N, l35)
	l36.LinkExit(location.S, l37)

	// Garden Shed
	l37.LinkExit(location.N, l36)

	// Dim Alley
	l38.LinkExit(location.N, l39)
	l38.LinkExit(location.S, l25)

	// Dim Alley
	l39.LinkExit(location.N, l40)
	l39.LinkExit(location.S, l38)

	// Bend in Dim Alley
	l40.LinkExit(location.S, l39)
	l40.LinkExit(location.W, l41)

	// Rogue's Den Entrance
	l41.LinkExit(location.E, l40)
	l41.LinkExit(location.D, l42)

	// Rogue's Den
	l42.LinkExit(location.U, l41)

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
	world.AddLocation(l16)
	world.AddLocation(l17)
	world.AddLocation(l18)
	world.AddLocation(l19)
	world.AddLocation(l20)
	world.AddLocation(l21)
	world.AddLocation(l22)
	world.AddLocation(l23)
	world.AddLocation(l24)
	world.AddLocation(l25)
	world.AddLocation(l26)
	world.AddLocation(l27)
	world.AddLocation(l28)
	world.AddLocation(l29)
	world.AddLocation(l30)
	world.AddLocation(l31)
	world.AddLocation(l32)
	world.AddLocation(l33)
	world.AddLocation(l34)
	world.AddLocation(l35)
	world.AddLocation(l36)
	world.AddLocation(l37)
	world.AddLocation(l38)
	world.AddLocation(l39)
	world.AddLocation(l40)
	world.AddLocation(l41)
	world.AddLocation(l42)

	// Some objects
	t1 := item.New(
		"A curious brass lattice",
		[]string{"LATTICE"},
		"This is a finely crafted, intricate lattice of fine brass wires forming a roughly ball shaped curiosity.",
		2,
	)
	t2 := item.New(
		"A small ball",
		[]string{"BALL"},
		"This is a small rubber ball.",
		18,
	)
	t3 := item.New(
		"An iron bound chest",
		[]string{"CHEST"},
		"This is vary stout wooden chest about 2 foot wide and 1 foot deep. Thick metal bands bind it.",
		89,
	)

	l1.Add(t1)
	l1.Add(t2)
	l1.Add(t3)
}
