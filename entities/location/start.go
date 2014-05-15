// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package location

import (
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"

	"math/rand"
)

// Start contains pointers to all of the available starting locations.
var start []*Start

// Start implements a starting location. That is a location where players can
// enter the world. It is simply a new type wrapping a Basic location.
type Start struct {
	Basic
}

// Register zero value instance of Start with the loader.
func init() {
	recordjar.RegisterUnmarshaler("start", &Start{})
}

// Unmarshal takes a recordjar.Record and allocates the data in it to the passed
// Start struct. It also adds a reference to the created location into the
// package scoped start slice.
func (s *Start) Unmarshal(d recordjar.Decoder) {
	defer func() {
		start = append(start, s)
	}()

	s.Basic.Unmarshal(d)
}

// GetStart return a random starting location.
func GetStart() *Start {
	return start[rand.Intn(len(start))]
}
