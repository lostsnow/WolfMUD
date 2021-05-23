// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package proc

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

type state struct {
	actor *Thing
	cmd   string
	word  []string
	buff  *bytes.Buffer
}

var filler = []string{"", "", ""}

func NewState(t *Thing, cmd string) *state {
	words := strings.Fields(strings.ToUpper(cmd))
	if len(words) < len(filler) {
		words = append(words, filler[len(words):]...)
	}
	return &state{
		t, words[0], words[1:], &bytes.Buffer{},
	}
}

func (s *state) Parse() {
	var start, end time.Time
	start = time.Now()

	switch s.cmd {
	case "":
		// Do nothing...
	case "L", "LOOK":
		s.Look()
	case
		"N", "NORTH", "NE", "NORTHEAST", "E", "EAST", "SE", "SOUTHEAST",
		"S", "SOUTH", "SW", "SOUTHWEST", "W", "WEST", "NW", "NORTHWEST",
		"UP", "DOWN":
		s.Move()
	case "EXAM", "EXAMINE":
		s.Examine()
	case "INV", "INVENTORY":
		s.Inv()
	case "DROP":
		s.Drop()
	case "GET":
		s.Get()
	case "TAKE":
		s.Take()
	case "PUT":
		s.Put()
	case "QUIT":
		s.buff.WriteString("Bye bye!\n")
	default:
		s.buff.WriteString("Eh?")
	}

	end = time.Now()
	if s.buff.Len() > 0 {
		s.buff.WriteByte('\n')
	}
	fmt.Printf("%s%s >", s.buff.String(), end.Sub(start))
}
