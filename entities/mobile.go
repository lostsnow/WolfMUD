package entities

import (
	"fmt"
	"strings"
)

type Mobile interface {
	Thing
	Inventory
	Parse(cmd string)
	Locate(l Location)
}

type mobile struct {
	thing
	inventory
	location Location
}

func NewMobile(name, alias, description string) (m Mobile) {
	return &mobile{
		thing: thing{name, alias, description},
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

func (m *mobile) Locate(l Location) {
	m.location = l
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
			handled = m.location.Command(what, cmd, args)
		}

	case "INVENTORY", "INV":
		handled = m.inv(what, args)
	}
	return handled
}

func (m *mobile) inv(what Thing, args []string) (handled bool) {
	if len(args) != 0 {
		return false
	}

	fmt.Println("You are currently carrying:")
	for _, v := range m.inventory.List(what) {
		fmt.Printf("\t%s\n", v.Name())
	}
	return true
}
