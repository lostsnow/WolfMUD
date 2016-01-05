// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

import (
	"sync"
)

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
	Attribute
	Name() string
}

type Description interface {
	Attribute
	Description() string
}

type Writing interface {
	Attribute
	Writing() string
}

type Vetoes interface {
	Attribute
	Check(...string) Veto
}

type Veto interface {
	Command() string
	Message() string
	Dump() []string
}

type Alias interface {
	Attribute
	HasAlias(string) bool
}

type Inventory interface {
	Attribute
	Add(Thing)
	Remove(Thing) Thing
	Search(string) Thing
	Contents() []Thing
	List() string
	sync.Locker
	LockID() uint64
	Crowded() bool
}

type Narrative interface {
	Attribute
	Add(Thing)
	Remove(Thing) Thing
	Search(string) Thing
	List() string
	ImplementsNarrative()
}

type Exits interface {
	Attribute
	Link(byte, Inventory)
	AutoLink(byte, Inventory)
	Unlink(byte)
	AutoUnlink(byte)
	List() string
	NormalizeDirection(string) string
	LeadsTo(string) Inventory
}

type Locate interface {
	Attribute
	Where() Inventory
	SetWhere(Inventory)
}

type Player interface {
	Attribute
	Write([]byte)
}
