package main

import (
	"wolfmud.org/entities"
	"net"
	"strconv"
)

func main() {
	world := []entities.Location{

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

	world[0].Add(t1)
	world[0].Add(t2)

	ln, err := net.Listen("tcp", ":4001")
	if err != nil {
		// handle error
	}

	println("WolfMUD server running...")

	n := 1

	for {
		println("Waiting for connection...")
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}
		println("Connection from: "+conn.RemoteAddr().String())
		postfix := strconv.Itoa(n)
		p := entities.NewPlayer("Player "+postfix,"PLAYER"+postfix,"This is Player "+postfix+".")
		n++
		world[0].Add(p)
		go p.Run(conn)
	}
}
