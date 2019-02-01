// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"math/rand"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Register marshaler for Start attribute.
func init() {
	internal.AddMarshaler((*Start)(nil), "start")
}

// registry is a list of all known/registered starting locations.
var registry []has.Start

// Start implements an attribute for tagging a Thing as a starting location.
type Start struct {
	Attribute
}

// Some interfaces we want to make sure we implement
var (
	_ has.Start = &Start{}
)

// NewStart returns a new Start attribute. When a new Start attribute is
// created it is also registered automatically.
//
// TODO: Implement starting locations that are  only usable by specific
// players. For example only dwarves should start in the dwarven home, or
// thieves in the thieves guild.
func NewStart() *Start {
	s := &Start{Attribute{}}
	registry = append(registry, s)
	return s
}

// FindStart searches the attributes of the specified Thing for attributes that
// implement has.Start returning the first match it finds or a *Start typed nil
// otherwise.
func FindStart(t has.Thing) has.Start {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Start); ok {
			return a
		}
	}
	return (*Start)(nil)
}

// Found returns false if the receiver is nil otherwise true.
func (s *Start) Found() bool {
	return s != nil
}

// Unmarshal is used to turn the passed data into a new Start attribute.
func (*Start) Unmarshal(data []byte) has.Attribute {
	return NewStart()
}

// Marshal returns a tag and []byte that represents the receiver.
func (*Start) Marshal() (tag string, data []byte) {
	return "start", data
}

func (s *Start) Dump() []string {
	return []string{DumpFmt("%p %[1]T", s)}
}

// Pick returns the Inventory of a randomly selected starting location.
func (*Start) Pick() has.Inventory {
	s := registry[rand.Intn(len(registry))]
	return FindInventory(s.Parent())
}

// Copy returns a copy of the Start receiver.
func (s *Start) Copy() has.Attribute {
	if s == nil {
		return (*Start)(nil)
	}
	return NewStart()
}

// Free deregisters the Start attribute when it is freed.
func (s *Start) Free() {
	for x, r := range registry {
		if s == r {
			copy(registry[x:], registry[x+1:])
			registry[len(registry)-1] = nil
			registry = registry[:len(registry)-1]
			break
		}
	}
}
