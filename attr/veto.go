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
	Attribute
	vetoes map[string]has.Veto
}

// Some interfaces we want to make sure we implement
var (
	_ has.Attribute = Vetoes()
	_ has.Vetoes    = Vetoes()
)

func Vetoes() *vetoes {
	return nil
}

func (*vetoes) New(veto ...has.Veto) *vetoes {
	vetoes := &vetoes{Attribute{}, make(map[string]has.Veto)}
	for _, v := range veto {
		vetoes.vetoes[v.Command()] = v
	}
	return vetoes
}

func (*vetoes) Find(t has.Thing) has.Vetoes {
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

func (v *vetoes) Check(cmd ...string) has.Veto {

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

type veto struct {
	cmd string
	msg string
}

// Some interfaces we want to make sure we implement
var (
	_ has.Veto = Veto()
)

func Veto() *veto {
	return nil
}

func (*veto) New(cmd string, msg string) *veto {
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
