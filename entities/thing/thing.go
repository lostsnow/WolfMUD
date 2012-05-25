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
// Due to the unique ID copies should not be made by assignment.
package thing

import (
	"log"
	"runtime"
	"strings"
	. "wolfmud.org/utils/UID"
)

// Interface should be implemented by all entities in WolfMUD. It provides
// everything with a name, description, aliases and a unique ID.
type Interface interface {
	Description() string
	IsAlias(alias string) bool
	IsAlso(thing Interface) bool
	Name() string
	UniqueId() UID
}

// The Thing type is a default implementation of the thing.Interface
type Thing struct {
	name        string
	description string
	aliases     []string
	uniqueId    UID
}

// New allocates a new Thing returning a pointer reference. A unique ID will
// be allocated. The aliases will all be stripped of leading and trailing
// whitespace converted to uppercase.
func New(name string, aliases []string, description string) *Thing {

	for i, a := range aliases {
		aliases[i] = strings.ToUpper(strings.TrimSpace(a))
	}

	t := &Thing{
		name:        name,
		aliases:     aliases,
		description: description,
		uniqueId:    <-Next,
	}

	log.Printf("Thing %d created: %s\n", t.uniqueId, t.name)

	runtime.SetFinalizer(t, Final)

	return t
}

func Final(t *Thing) {
	log.Printf("+++ Thing %d finalized: %s +++\n", t.uniqueId, t.name)
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

// IsAlso tests two things to see if one of them 'is also' the other - hence the
// functions name.
//
// WolfMUD uses a lot of Interfaces and embedded types. So we may be comparing,
// for example, a Player with a Mobile. However this causes issues:
//
//	- Mobile and Player are not the same types
//	- They can have different interfaces
//	- Pointers to a Mobile embeded in a Player will be different (of course)
//
// So to make things easy we have the unique ID and can use either of:
//
//	thisPlayer.IsAlso(thisMobile)
//	thisPlayer.UniqueId() == thisMobile.UniqueId()
func (t *Thing) IsAlso(thing Interface) bool {
	return t.uniqueId == thing.UniqueId()
}

// Name returns the name given to a Thing.
func (t *Thing) Name() string {
	return t.name
}

// UniqueId returns the unique ID of a Thing.
func (t *Thing) UniqueId() UID {
	return t.uniqueId
}