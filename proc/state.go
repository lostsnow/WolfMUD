// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package proc

import (
	"io"
	"strings"
)

type state struct {
	actor  *Thing
	cmd    string
	word   []string
	out    io.Writer
	buff   *strings.Builder
	prompt []byte
}

var (
	World   map[string]*Thing
	filler  = []string{"", "", ""}
	newline = []byte("\n")
)

func NewState(out io.Writer, t *Thing) *state {
	return &state{
		actor: t, out: out, buff: &strings.Builder{}, prompt: []byte(">"),
	}
}

func (s *state) Parse(input string) {
	if input == "\n" || input == "" {
		s.out.Write(s.prompt)
		return
	}
	s.parse(input)
	s.out.Write([]byte(s.buff.String()))
	s.buff.Reset()
	s.out.Write(newline)
	s.out.Write(s.prompt)
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
