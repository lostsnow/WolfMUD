package thing

import (
	. "wolfmud.org/utils/UID"
)

type Interface interface {
	Description() string
	IsAlias(alias string) bool
	IsAlso(thing Interface) bool
	Name() string
	UniqueId() UID
}

type Thing struct {
	name        string
	description string
	aliases     []string
	uniqueId    UID
}

func (t *Thing) Name() string {
	return t.name
}

func (t *Thing) Description() string {
	return t.description
}

func (t *Thing) IsAlias(alias string) bool {
	for _, a := range t.aliases {
		if a == alias {
			return true
		}
	}
	return false
}

func (t *Thing) UniqueId() UID {
	return t.uniqueId
}

func (t *Thing) IsAlso(thing Interface) bool {
	return t.uniqueId == thing.UniqueId()
}

func New(name string, aliases []string, description string) *Thing {
	return &Thing{
		name:        name,
		aliases:     aliases,
		description: description,
		uniqueId:    <-Next,
	}
}
