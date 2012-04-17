package entities

import (
	"fmt"
)

type Inventory interface {
	// Commander
	Add(t Thing)
	Remove(alias string, occurance int) (t Thing)
}

type inventory struct {
	content map[string][]Thing
}

func NewInventory() Inventory {
	return &inventory{
	// content has lazy initialisation, see Add and Remove
	}
}

func (i *inventory) delegate(what Thing, cmd string, args []string) (handled bool) {
	// An inventory delegates to everything in it but handles nothing itself
ABORT:
	for _, alias := range i.content {
		for _, object := range alias {

			// Don't process ourself at a location - gets recursive!
			if what == object {
				continue
			}

			if _, ok := object.(Commander); ok {
				handled = object.Command(what, cmd, args)
				if handled {
					break ABORT
				}
			}
		}
	}
	return handled
}

func (i *inventory) Add(t Thing) {
	if i.content == nil {
		i.content = make(map[string][]Thing)
	}
	i.content[t.Alias()] = append(i.content[t.Alias()], t)
}

func (i *inventory) Remove(alias string, occurance int) (t Thing) {

	qty := len(i.content[alias])

	switch {
	case occurance == 0:
		fmt.Printf("You can't drop nothing of something!\n")

	case occurance > qty:
		fmt.Printf("There are not that many '%s', you can only find %d!\n", alias, qty)

	default:
		occurance--
		t = i.content[alias][occurance]
		i.content[alias] = append(i.content[alias][:occurance], i.content[alias][occurance+1:]...)

		// If we started with 1 we now have 0 so delete bucket
		if qty == 1 {
			delete(i.content, alias)
		}

		// If inventory now empty drop it
		if len(i.content) == 0 {
			i.content = nil
		}
	}

	return
}

func (i *inventory) List(ommit Thing) (list string) {
	for _, alias := range i.content {
		for _, object := range alias {
			if object == ommit {
				continue
			}
			if _, ok := object.(Player); ok {
				list = "You can see " + object.Name() + " here.\n" + list
			} else {
				list += "You can see " + object.Name() + " here.\n"
			}
		}
	}
	return list
}
