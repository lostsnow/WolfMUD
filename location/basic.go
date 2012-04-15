package location

import (
	"fmt"
)

/*
	The direction type is defined so that functions can take a direction for the
	exits array index. Then a user can only pass one of the valid defined
	direction values.  Note that this type is not exported which would allow user
	defined and probably invalid values to be created.
*/
type direction uint8

/*
	Direction constants of type direction used to index exits array. Only these valid
	constants can be used. Note that the constants ARE exported while the type is
	not. This is valid as a user will refer to the constants and not the type.
*/
const (
	N, NORTH direction = iota, iota
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

var directionNames = [10]string{
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

func (from *Basic) SetExit(d direction, l Location) {
	from.exits[d] = l
}

func (from *Basic) Move(d direction) (to Location) {
	if to = from.exits[d]; to != nil {
		fmt.Printf("You go %s.\n", directionNames[d])
		to.Look(nil)
	} else {
		fmt.Printf("You can't go %s from here!\n", directionNames[d])
		to = from
	}
	return
}

func (from *Basic) Look(args []string) (handled bool) {
	if len(args) > 0 {
		return
	}

	fmt.Println("")
	fmt.Println(from.name)
	fmt.Println(from.description)
	from.Exits()
	fmt.Println()
	return true
}

func (from *Basic) Exits() {
	found := false

	for d, l := range from.exits {
		if l != nil {
			if found == false {
				fmt.Print("Exits you can see are:")
				found = true
			}
			fmt.Print(" ", directionNames[d])
		}
	}
	if found == false {
		fmt.Print("There are no obvious exits!")
	}
	fmt.Println("")
}

func (from *Basic) Command(c string, args []string) (handled bool) {
	switch c {
	case `LOOK`:
		handled = from.Look(args)
	case `NORTH`,`N`:
		from.Move(NORTH)
		handled = true
	case `SOUTH`,`S`:
		from.Move(SOUTH)
		handled = true
	case `EAST`,`E`:
		from.Move(EAST)
		handled = true
	case `WEST`,`W`:
		from.Move(WEST)
		handled = true
	}
	return
}
