/*
	Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.

	Use of this source code is governed by the license in the LICENSE file
	included with the source code.
*/

package command

import (
	"strings"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/responder"
)

type Interface interface {
	Process(*Command) (handled bool)
}

type Command struct {
	Issuer thing.Interface
	Verb   string
	Nouns  []string
	Target *string
}

func New(issuer thing.Interface, input string) *Command {
	words := strings.Split(strings.ToUpper(input), ` `)

	cmd := Command{}

	cmd.Issuer = issuer
	cmd.Verb = words[0]
	cmd.Nouns = words[1:]

	if len(words) > 1 {
		cmd.Target = &words[1]
	}

	return &cmd
}

func (c *Command) Respond(format string, any ...interface{}) {
	if i, ok := c.Issuer.(responder.Interface); ok {
		i.Respond(format, any...)
	}
}
