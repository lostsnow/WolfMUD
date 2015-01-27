// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"strings"
)

type vetoes struct {
	attribute
	vetoes map[string]has.Veto
}

// Some interfaces we want to make sure we implement
var _ has.Attribute = &vetoes{}
var _ has.Vetoes = &vetoes{}

func NewVetoes(veto ...has.Veto) *vetoes {
	vetoes := &vetoes{attribute{}, make(map[string]has.Veto)}
	for _, v := range veto {
		vetoes.vetoes[v.Command()] = v
	}
	return vetoes
}

func FindVeto(t has.Thing) (v has.Vetoes) {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Vetoes); ok {
			return a
		}
	}
	return nil
}

func (v *vetoes) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T %d vetoes:", v, len(v.vetoes)))
	for _, veto := range v.vetoes {
		for _, line := range veto.Dump() {
			buff = append(buff, DumpFmt("%s", line))
		}
	}
	return buff
}

func (v *vetoes) Check(cmd string) has.Veto {
	veto, _ := v.vetoes[cmd]
	return veto
}

type veto struct {
	cmd string
	msg string
}

// Some interfaces we want to make sure we implement
var _ has.Veto = &veto{}

func NewVeto(cmd string, msg string) *veto {
	return &veto{strings.ToUpper(cmd), msg}
}

func (v *veto) Dump() (buff []string) {
	return append(buff, DumpFmt("%p %[1]T %q:%q", v, v.Command(), v.Message()))
}

func (v *veto) Command() string {
	return v.cmd
}

func (v *veto) Message() string {
	return v.msg
}
