package entities

import (
	"fmt"
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
	handled := m.Command(NewCmd(m, cmd))
	if handled == false {
		fmt.Printf("Eh? %s?\n\n", cmd)
	}
}

func (m *mobile) Locate(l Location) {
	m.location = l
}


func (m *mobile) Command(c Cmd) (handled bool) {
	switch c.Verb() {
	default:
		handled = m.thing.Command(c)

		if handled == false && c.What() == m {
			handled = m.inventory.delegate(c)
		}

		// If we are handling commands for ourself can our environment handle it?
		if handled == false && c.What() == m {
			handled = m.location.Command(c)
		}

	case "INVENTORY", "INV":
		handled = m.inv(c)
	}
	return handled
}

func (m *mobile) inv(c Cmd) (handled bool) {
	if c.Target() != nil {
		return false
	}

	fmt.Println("You are currently carrying:")
	for _, v := range m.inventory.List(c.What()) {
		fmt.Printf("\t%s\n", v.Name())
	}
	return true
}
