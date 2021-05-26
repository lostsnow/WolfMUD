// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package proc

import (
	"fmt"
	"strings"
	"time"
)

type state struct {
	actor *Thing
	cmd   string
	word  []string
	buff  *strings.Builder
}

var World map[string]*Thing
var filler = []string{"", "", ""}

func NewState(t *Thing, cmd string) *state {
	words := strings.Fields(strings.ToUpper(cmd))
	if len(words) < len(filler) {
		words = append(words, filler[len(words):]...)
	}
	return &state{
		t, words[0], words[1:], &strings.Builder{},
	}
}

func (s *state) Parse() {
	var start, end time.Time
	start = time.Now()

	if command, ok := commands[s.cmd]; ok {
		command(s)
	} else {
		s.Msg("Eh?")
	}

	end = time.Now()
	if s.buff.Len() > 0 {
		s.buff.WriteByte('\n')
	}
	fmt.Printf("%s%s >", s.buff.String(), end.Sub(start))
}

func (s *state) Msg(text ...string) {
	for _, t := range text {
		s.buff.WriteString(t)
	}
}
