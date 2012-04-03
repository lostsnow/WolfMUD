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

	var room location.Location

	rooms := [...]location.Location{
		location.NewBasic("Room one", "You are in the first room."),
		location.NewBasic("Room two", "You are in the second room."),
	}

	rooms[0].SetNorth(rooms[1])
	rooms[1].SetSouth(rooms[0])

	room = rooms[0]
	room.Look()
	room = room.North()
	room = room.North()
	room = room.South()
	room = room.South()

}
