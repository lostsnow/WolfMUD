package main

import (
	"fmt"
	"wolfmud.org/entities"
)

func main() {
	fmt.Println("\n+++ HELLO WORLD +++")

	world := [...]entities.Location{

		entities.NewLocation("Fireplace", "FIREPLACE", "You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. To the south the common room extends and east the common room leads to the tavern entrance."),

		entities.NewLocation("Common Room", "COMMONROOM", "You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away."),

		entities.NewLocation("Tavern Bar", "TAVERNBAR", "You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here."),

		entities.NewLocation("Tavern entrance", "TAVERNENTRANCE", "You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street."),
	}

	// Fireplace
	world[0].LinkExit(entities.E, world[3])
	world[0].LinkExit(entities.SE, world[2])
	world[0].LinkExit(entities.S, world[1])

	// Common room
	world[1].LinkExit(entities.N, world[0])
	world[1].LinkExit(entities.NE, world[3])
	world[1].LinkExit(entities.E, world[2])

	// Tavern Bar
	world[2].LinkExit(entities.N, world[3])
	world[2].LinkExit(entities.W, world[1])
	world[2].LinkExit(entities.NW, world[0])

	// Tavern Entrance
	world[3].LinkExit(entities.S, world[2])
	world[3].LinkExit(entities.SW, world[1])
	world[3].LinkExit(entities.W, world[0])

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

	// Some mobiles
	m1 := entities.NewMobile(
		"Diddymus",
		"DIDDYMUS",
		"An adventurer like yourself.",
	)
	m2 := entities.NewMobile(
		"Tass",
		"TASS",
		"An adventurer like yourself.",
	)

	// Put lattice into the world
	world[0].Add(t1)

	// Put Tass into the world
	m2.Locate(world[0])
	world[0].Add(m2)

	// Put ball into Diddymus' inventory, then add to world
	m1.Add(t2)
	m1.Locate(world[0])
	world[0].Add(m1)

	m1.Command(m1, "LOOK", nil)

	m1.Parse("look")
	m1.Parse("look lattice")
	m1.Parse("look tass")
	m1.Parse("inventory")
	m1.Parse("inv")
	m1.Parse("look ball")
	m1.Parse("examine ball")
	m1.Parse("west")
	m1.Parse("south")
	m1.Parse("n")

	println("\n+++ GOODBYE WORLD +++\n")
}
