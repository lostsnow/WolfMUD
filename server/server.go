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

	var l location.Location

	world := [...]location.Location{
		location.NewBasic("Room one", "You are in the first location."),
		location.NewBasic("Room two", "You are in the second location."),
	}

	world[0].SetExit("North", world[1])
	world[1].SetExit("South", world[0])

	l = world[0]
	l.Look()
	l = l.Move("North")
	l = l.Move("North")
	l = l.Move("South")
	l = l.Move("South")

}
