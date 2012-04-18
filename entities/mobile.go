package entities

import (
	"fmt"
	"strings"
)

type Mobile interface {
	Thing
	Inventory
	Parse(cmd string)
}

type mobile struct {
	thing
	inventory
}

func NewMobile(name, alias, description string, location Location) (m Mobile) {
	return &mobile{
		thing: thing{name, alias, description, location},
	}
}

func (m *mobile) Parse(cmd string) {
	fmt.Printf("\n> %s\n", cmd)
	words := strings.Split(strings.ToUpper(cmd), " ")
	handled := m.Command(m, words[0], words[1:])
	if handled == false {
		fmt.Printf("Eh? %s?\n\n", cmd)
	}
}

func (m *mobile) Command(what Thing, cmd string, args []string) (handled bool) {
	switch cmd {
	default:
		handled = m.thing.Command(what, cmd, args)

		if handled == false && what == m {
			handled = m.inventory.delegate(what, cmd, args)
		}

		// If we are handling commands for ourself can our environment handle it?
		if handled == false && what == m {
			handled = m.thing.location.Command(what, cmd, args)
		}

	case "INVENTORY", "INV":
		handled = m.Inventory(what, args)
	}
	return handled
}

func (m *mobile) Inventory(what Thing, args []string) (handled bool) {
	if len(args) != 0 {
		return false
	}

	fmt.Println("You are currently carrying:")
	for _, v := range m.inventory.List(what) {
		fmt.Printf("\t%s\n", v.Name())
	}
	return true
}

