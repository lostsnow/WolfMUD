package main

import (
	"fmt"
	"wolfmud.org/location"
)

func main() {
	fmt.Println(`
-----------------------
Starting WolfMUD server
-----------------------`)

	world := [...]location.Location{
		location.NewBasic("Fireplace", "You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. To the south the common room extends and east the common room leads to the tavern entrance."),
		location.NewBasic("Common Room", "You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away."),
		location.NewBasic("Tavern Bar", "You standing at the bar. Behind which you can see various sized and shaped bottles. Looking at the contents you decide an abstract painter would get lots of colourful inspirations after a long night here."),
		location.NewBasic("Tavern entrance", "You are in the entryway to the Dragon's Breath tavern. To the west you can see an inviting fireplace, while south an even more inviting bar. Eastward a door leads out into the street."),
	}

	// Fireplace
	world[0].SetExit(location.E, world[3])
	world[0].SetExit(location.S, world[1])
	world[0].SetExit(location.SE, world[2])

	// Common room
	world[1].SetExit(location.N, world[0])
	world[1].SetExit(location.E, world[2])
	world[1].SetExit(location.NE, world[3])

	// Tavern Bar
	world[2].SetExit(location.N, world[3])
	world[2].SetExit(location.W, world[1])
	world[2].SetExit(location.NW, world[0])

	// Tavern Entrance
	world[3].SetExit(location.S, world[2])
	world[3].SetExit(location.W, world[0])
	world[3].SetExit(location.SW, world[1])

}
