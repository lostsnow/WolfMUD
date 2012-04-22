package entities

import (
	"fmt"
)

type Player interface{
	Mobile
	Responder
}

type player struct {
	mobile
	responder
}

func NewPlayer(name, alias, description string) Player {
	m := NewMobile(name, alias, description).(*mobile)
	return &player{
		mobile: *m,
	}
}

func (p *player) Parse(input string) {
	fmt.Printf("\n> %s\n", input)
	handled := p.Process(NewCommand(p, input))
	if handled == false {
		fmt.Printf("Eh? %s?\n\n", input)
	}
}
