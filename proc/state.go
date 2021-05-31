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

func NewState(t *Thing) *state {
	return &state{
		actor: t, buff: &strings.Builder{},
	}
}

func (s *state) Parse(input string) {
	var start, end time.Time
	start = time.Now()

	if input != "\n" {
		s.parse(input)
		fmt.Println(s.buff.String())
		s.buff.Reset()
	}

	end = time.Now()
	fmt.Printf("%s >", end.Sub(start))
}

func (s *state) parse(input string) {
	s.word = strings.Fields(strings.ToUpper(input))
	if len(s.word) < len(filler) {
		s.word = append(s.word, filler[len(s.word):]...)
	}
	s.cmd, s.word = s.word[0], s.word[1:]

	if command, ok := commands[s.cmd]; ok {
		command(s)
	} else {
		s.Msg("Eh?")
	}
}

func (s *state) Msg(text ...string) {
	for _, t := range text {
		s.buff.WriteString(t)
	}
}
