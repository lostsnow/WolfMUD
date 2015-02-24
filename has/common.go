// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

type Thing interface {
	Add(...Attribute)
	Remove(...Attribute)
	Attrs() []Attribute
	Dump() []string
}

type Attribute interface {
	Parent() Thing
	SetParent(Thing)
	Dump() []string
}

type Name interface {
	Name() string
}

type Description interface {
	Description() string
}

type Writing interface {
	Writing() string
}

type Vetoes interface {
	Check(...string) Veto
}

type Veto interface {
	Command() string
	Message() string
	Dump() []string
}

type Alias interface {
	HasAlias(string) bool
}

type Inventory interface {
	Add(Thing)
	Remove(Thing) Thing
	Search(string) Thing
	Contains(Thing) bool
	List() []Thing
	Contents() string
}

type Narrative interface {
	Add(Thing)
	Remove(Thing) Thing
	Search(string) Thing
	Contains(Thing) bool
	Contents() string
	ImplementsNarrative()
}

type Exits interface {
	Link(uint8, Thing)
	Unlink(uint8)
	List() string
	Place(Thing)
	Move(Thing, string) (string, bool)
}

type Locate interface {
	Where() Thing
	SetWhere(Thing)
}
