package location

import (
	"fmt"
)

type Basic struct {
	north       Location
	south       Location
	name        string
	description string
}

func NewBasic(n, d string) (b *Basic) {
	return &Basic{
		name:        n,
		description: d,
	}
}

func (from *Basic) SetNorth(l Location) {
	from.north = l
}

func (from *Basic) SetSouth(l Location) {
	from.south = l
}

func (from *Basic) North() (to Location) {
	if from.north == nil {
		fmt.Println("You can't go north from here!")
		to = from
	} else {
		fmt.Println("You go north.")
		to = from.north
		to.Look()
	}
	return
}

func (from *Basic) South() (to Location) {
	if from.south == nil {
		fmt.Println("You can't go south from here!")
		to = from
	} else {
		fmt.Println("You go south.")
		to = from.south
		to.Look()
	}
	return
}

func (from *Basic) Look() {
	fmt.Println("")
	fmt.Println(from.name)
	fmt.Println(from.description)
	from.Exits()
	fmt.Println()
}

func (from *Basic) Exits() {

	fmt.Print("Exits you can see are:")
	if from.north != nil {
		fmt.Print(" North")
	}
	if from.south != nil {
		fmt.Print(" South")
	}
	fmt.Println("")

}
