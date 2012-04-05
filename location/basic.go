package location

import (
	"fmt"
)

type Basic struct {
	exits       map[string]Location
	name        string
	description string
}

func NewBasic(n, d string) (b *Basic) {
	return &Basic{
		exits:       map[string]Location{},
		name:        n,
		description: d,
	}
}

func (from *Basic) SetExit(d string, l Location) {
	from.exits[d] = l
}

func (from *Basic) Move(d string) (to Location) {
	if l, ok := from.exits[d]; ok {
		fmt.Printf("You go %s.\n", d)
		to = l
		to.Look()
	} else {
		fmt.Printf("You can't go %s from here!\n", d)
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
	fmt.Print("Exits you can see are:")
	for e := range from.exits {
		fmt.Print(" ", e)
	}
	fmt.Println("")
}
