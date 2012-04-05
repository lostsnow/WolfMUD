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
		location.NewBasic("Room one", "You are in the first location."),
		location.NewBasic("Room two", "You are in the second location."),
		location.NewBasic("Room three", "You are in the third location."),
	}

	world[0].SetExit(location.N, world[1])
	world[0].SetExit(location.SE, world[2])
	world[1].SetExit(location.SOUTH, world[0])

	l := world[0]
	l.Look()
	l = l.Move(location.N)
	l = l.Move(location.NORTH)
	l = l.Move(location.S)
	l = l.Move(location.SOUTH)
	l = l.Move(location.SE)

}
