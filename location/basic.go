package location

import (
	"fmt"
)

/*
	The Exit type is defined so that functions can take an Exit for the exits
	array index. Then a user can only pass one of the defined valid values in.
	Note that this type is not exported which would allow user defined and
	possibly invalid values to be created.
*/
type exit uint8

/*
	Exit constants of type exit used to index exits array. Only these valid
	types can be used. Note that the constants ARE exported while the type is
	not. This is valid as a user will refer to the constant and not the type.
*/
const (
	N, NORTH exit = iota, iota
	NE, NORTHEAST
	E, EAST
	SE, SOUTHEAST
	S, SOUTH
	SW, SOUTHWEST
	W, WEST
	NW, NORTHWEST
	U, UP
	D, DOWN
)

var exitNames = [10]string{
	N:  "North",
	NE: "Northeast",
	E:  "East",
	SE: "Southeast",
	S:  "South",
	SW: "Southwest",
	W:  "West",
	NW: "Northwest",
	U:  "Up",
	D:  "Down",
}

type Basic struct {
	exits       [10]Location
	name        string
	description string
}

func NewBasic(n, d string) (b *Basic) {
	return &Basic{
		name:        n,
		description: d,
	}
}

func (from *Basic) SetExit(d exit, l Location) {
	from.exits[d] = l
}

func (from *Basic) Move(d exit) (to Location) {
	if to = from.exits[d]; to != nil {
		fmt.Printf("You go %s.\n", exitNames[d])
		to.Look()
	} else {
		fmt.Printf("You can't go %s from here!\n", exitNames[d])
		to = from
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
	found := false

	for d, l := range from.exits {
		if l != nil {
			if found == false {
				fmt.Print("Exits you can see are:")
				found = true;
			}
			fmt.Print(" ", exitNames[d])
		}
	}
	if found == false {
		fmt.Print("There are no obvious exits!")
	}
	fmt.Println("")
}
