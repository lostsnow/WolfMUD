// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package thing implements the base type of all entities in WolfMUD. Each Thing
// is created with a unique ID. This aids identifying when two Things are the
// same Thing no matter which Interface the Thing is 'seen through' or which
// embed type is in use. Things can easily be compared in one of two ways:
//
//	thing1.IsAlso(thing2)
//
//	thing1.UniqueId() == thing2.UniqueId()
//
// Due to the unique ID copies should not be made by assignment unless a new
// unique ID is allocated or the assignment is very temporary.
package thing

import (
	"log"
	"strings"
	. "wolfmud.org/utils/uid"
)

// Interface should be implemented by all entities in WolfMUD. It provides
// everything with a name, description, aliases and a unique ID.
type Interface interface {
	Description() string
	IsAlias(alias string) bool
	Aliases() []string
	IsAlso(thing Interface) bool
	Lock()
	Name() string
	UniqueId() UID
	Unlock()
}

// Thing type is a default implementation of the thing.Interface
type Thing struct {
	name        string
	description string
	aliases     []string
	uniqueId    UID
	mutex       chan bool
}

// New allocates a new Thing, returning a pointer reference to it. A unique ID
// will be allocated automatically. The aliases will all be stripped of leading
// and trailing whitespace then converted to uppercase.
func New(name string, aliases []string, description string) *Thing {

	t := &Thing{
		name:        name,
		aliases:     make([]string, len(aliases)),
		description: description,
		uniqueId:    <-Next,
		mutex:       make(chan bool, 1),
	}

	for i, a := range aliases {
		t.aliases[i] = strings.ToUpper(strings.TrimSpace(a))
	}

	log.Printf("Thing %d created: %s\n", t.uniqueId, t.name)

	return t
}

// Description returns the description for a Thing.
func (t *Thing) Description() string {
	return t.description
}

// IsAlias returns true if the passed string is one of the Thing's aliases,
// otherwise false. The comparison is case insensitive. The passed alias to be
// checked will be trimmed of leading and trailing whitespace and uppercased
// before the comparison.
//
// This method is not optimised as we usually expect only 2 or 3 aliases. If
// there is the need for a HUGE number of aliases we may want to store hashes to
// save memory and/or use a map with the aliases as the keys and simply test if
// the map element exists.
//
// Currently we brute force using a for/range and bail early when a match is
// found.
func (t *Thing) IsAlias(alias string) bool {

	alias = strings.ToUpper(strings.TrimSpace(alias))

	for _, a := range t.aliases {
		if a == alias {
			return true
		}
	}
	return false
}

func (t *Thing) Aliases() (a []string) {
	a = make([]string, len(t.aliases))
	copy(a, t.aliases)
	return
}

// IsAlso tests two Things to see if one of them 'is also' the other - hence the
// function's name.
//
// WolfMUD uses a lot of Interfaces and embedded types. So we may be comparing,
// for example, a Player with a Mobile. However this causes issues:
//
//	- Mobile and Player are not the same types
//	- They can have different interfaces
//	- Pointers to a Mobile embedded in a Player will be different (of course)
//
// So to make things easy we have the unique ID and can use either of:
//
//	thisPlayer.IsAlso(thisMobile)
//	thisPlayer.UniqueId() == thisMobile.UniqueId()
//
// The first example using IsAlso tends to make the code easier to read.
func (t *Thing) IsAlso(thing Interface) bool {
	return t.uniqueId == thing.UniqueId()
}

// Lock is a blocking channel lock. It is unlocked by calling Unlock. Unlock
// should only be called when the lock is held via a successful Lock call. The
// reason for the method instead of making the lock in the struct public - you
// cannot access struct properties directly through the Interface.
func (t *Thing) Lock() {
	t.mutex <- true
}

// Unlock unlocks a locked Thing. See Lock method for details.
func (t *Thing) Unlock() {
	<-t.mutex
}

// Name returns the name given to a Thing.
func (t *Thing) Name() string {
	return t.name
}

// UniqueId returns the unique ID of a Thing.
func (t *Thing) UniqueId() UID {
	return t.uniqueId
}
