package inventory

import (
	"wolfmud.org/entities/thing"
)

type Interface interface {
	Add(thing thing.Interface)
	Remove(thing thing.Interface)
	List(ommit thing.Interface) ([]thing.Interface)
}

type Inventory struct {
	contents []thing.Interface
}

func New() *Inventory {
	return &Inventory{}
}

func (i *Inventory) Add(thing thing.Interface) {
	i.contents = append(i.contents, thing)
}

func (i *Inventory) Remove(thing thing.Interface) {
	for index, t := range i.contents {
		if t == thing {
			i.contents = append(i.contents[:index], i.contents[index+1:]...)
		}
	}
}

func (i *Inventory) List(ommit thing.Interface) (list []thing.Interface) {

	for _, thing := range i.contents {
		if thing == ommit {
			continue
		}
		list = append(list, thing)
	}

	return
}
