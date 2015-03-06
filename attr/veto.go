// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"strings"
)

type Vetoes struct {
	Attribute
	vetoes map[string]has.Veto
}

// Some interfaces we want to make sure we implement
var (
	_ has.Vetoes = &Vetoes{}
)

func NewVetoes(veto ...has.Veto) *Vetoes {
	vetoes := &Vetoes{Attribute{}, make(map[string]has.Veto)}
	for _, v := range veto {
		vetoes.vetoes[v.Command()] = v
	}
	return vetoes
}

func FindVetoes(t has.Thing) has.Vetoes {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Vetoes); ok {
			return a
		}
	}
	return nil
}

func (v *Vetoes) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T %d vetoes:", v, len(v.vetoes)))
	for _, veto := range v.vetoes {
		for _, line := range veto.Dump() {
			buff = append(buff, DumpFmt("%s", line))
		}
	}
	return buff
}

func (v *Vetoes) Check(cmd ...string) has.Veto {

	// For single checks we can take a shortcut
	if len(cmd) == 1 {
		veto, _ := v.vetoes[cmd[0]]
		return veto
	}

	// For multiple checks return the first that is vetoed
	for _, cmd := range cmd {
		if veto, _ := v.vetoes[cmd]; veto != nil {
			return veto
		}
	}
	return nil
}

type Veto struct {
	cmd string
	msg string
}

// Some interfaces we want to make sure we implement
var (
	_ has.Veto = &Veto{}
)

func NewVeto(cmd string, msg string) *Veto {
	return &Veto{strings.ToUpper(cmd), msg}
}

func (v *Veto) Dump() (buff []string) {
	return append(buff, DumpFmt("%p %[1]T %q:%q", v, v.Command(), v.Message()))
}

func (v *Veto) Command() string {
	return v.cmd
}

func (v *Veto) Message() string {
	return v.msg
}
