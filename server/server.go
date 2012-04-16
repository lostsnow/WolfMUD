package main

import (
	"fmt"
	// "time"
	// "wolfmud.org/location"
	// "wolfmud.org/player"
	"wolfmud.org/entities"
)

func main() {
	fmt.Println(`
-----------------------
Starting WolfMUD server
-----------------------`)

	world := [...]entities.Location{
		entities.NewLocation("Fireplace", "You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. To the south the common room extends and east the common room leads to the tavern entrance.", nil),
		entities.NewLocation("Common Room", "You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away.", nil),
		entities.NewLocation("Tavern Bar", "You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here.", nil),
		entities.NewLocation("Tavern entrance", "You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street.", nil),
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

	/*
		p := player.New("Diddymus", world[0])
		p.Parse("LOOK")
		p.Parse("S")
		p.Parse("E")
		p.Parse("N")
		p.Parse("W")
		p.Parse("Say The quick brown fox jumps over the lazy dog but that's not enough for a line length of over 80 characters!")

		time.Sleep(2 * time.Second)
	*/

	p1 := entities.NewPlayer("Diddymus", "An adventurer like yourself", world[0])
	p2 := entities.NewPlayer("Tass", "An adventurer like yourself", world[0])

	world[0].Add(p2)

	world[0].Add(p1)
	p1.Command(p1, "LOOK", nil)
	p1.Command(p1, "LOOK", []string{"TASS"})

	p1.Command(p1, "SOUTH", nil)
	p1.Command(p1, "EAST", nil)
	p1.Command(p1, "NORTH", nil)
	p1.Command(p1, "W", nil)

}
