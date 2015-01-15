// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"strings"
)

type veto struct {
	parent
	vetoes map[string]string
}

func NewVeto(vetos [][2]string) *veto {
	v := make(map[string]string)
	for _, vs := range vetos {
		v[strings.ToUpper(vs[0])] = vs[1]
	}
	return &veto{parent{}, v}
}

func FindVeto(t has.Thing) has.Veto {

	compare := func(a has.Attribute) (ok bool) { _, ok = a.(has.Veto); return }

	if t := t.FindAttr(compare); t != nil {
		return t.(has.Veto)
	}
	return nil
}

func (v *veto) Dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T %d vetoes:", v, len(v.vetoes)))
	for cmd, msg := range v.vetoes {
		buff = append(buff, DumpFmt("  %q: %q", cmd, msg))
	}
	return buff
}

func (v *veto) Check(cmd string) string {
	if v, found := v.vetoes[cmd]; found {
		return v
	}
	return ""
}
