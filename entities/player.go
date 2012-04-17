package entities

import (
	"fmt"
	"strings"
)

type Player interface {
	Thing
	Inventory
	Parse(cmd string)
}

type player struct {
	thing
	inventory
}

func NewPlayer(name, alias, description string, location Location) (p Player) {
	return &player{
		thing: thing{name, alias, description, location},
	}
}

func (p *player) Parse(cmd string) {
	fmt.Printf("\n> %s\n", cmd)
	words := strings.Split(strings.ToUpper(cmd), " ")
	handled := p.Command(p, words[0], words[1:])
	if handled == false {
		fmt.Printf("Eh? %s?\n\n", cmd)
	}
}

func (p *player) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	default:
		handled = p.thing.Command(what, cmd, args)

		if handled == false && what == p {
			handled = p.inventory.delegate(what, cmd, args)
		}

		// If we are handling commands for ourself can our environment handle it?
		if handled == false && what == p {
			handled = p.thing.location.Command(what, cmd, args)
		}

	case "INVENTORY", "INV":
		handled = p.Inventory(what, args)
	}
	return handled
}

func (p *player) Inventory(what Thing, args []string) (handled bool) {
	if len(args) != 0 {
		return false
	}

	fmt.Println("You are currently carrying:")
	for _, v := range p.inventory.List(what) {
		fmt.Printf("\t%s\n", v.Name())
	}
	return true
}

