package entities

import (
	"strings"
)

type Commander interface {
	Command(Cmd) (handled bool)
}

type Cmd interface {
	What() Thing
	Verb() string
	Nouns() []string
	Target() *string
}

type cmd struct {
	what   Thing
	verb   string
	nouns  []string
}

func NewCmd(what Thing, input string) (Cmd) {
	words := strings.Split(strings.ToUpper(input), " ")
	return &cmd{what, words[0], words[1:]}
}

func (c *cmd) What() Thing {
	return c.what
}

func (c *cmd) Verb() string {
	return c.verb
}

func (c *cmd) Nouns() []string {
	return c.nouns
}

func (c *cmd) Target() (target *string) {
	if len(c.nouns) > 0 {
		target = &c.nouns[0]
	}
	return
}
