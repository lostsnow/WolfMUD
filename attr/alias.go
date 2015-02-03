// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"strconv"
	"strings"
)

type alias struct {
	attribute
	aliases map[string]struct{}
}

// Some interfaces we want to make sure we implement
var (
	_ has.Attribute = Alias()
	_ has.Alias     = Alias()
)

func Alias() *alias {
	return nil
}

func (*alias) New(a ...string) *alias {
	aliases := make(map[string]struct{}, len(a))
	for _, a := range a {
		aliases[strings.ToUpper(a)] = struct{}{}
	}
	return &alias{attribute{}, aliases}
}

func (*alias) Find(t has.Thing) has.Alias {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Alias); ok {
			return a
		}
	}
	return nil
}

func (a *alias) Dump() []string {
	buff := []byte{}
	for a := range a.aliases {
		buff = append(buff, ", "...)
		buff = strconv.AppendQuote(buff, a)
	}
	if len(buff) > 0 {
		buff = buff[2:]
	}
	return []string{DumpFmt("%p %[1]T %d aliases: %s", a, len(a.aliases), buff)}
}

func (a *alias) HasAlias(alias string) (found bool) {
	_, found = a.aliases[alias]
	return
}
