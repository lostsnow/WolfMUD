// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func Examine(t has.Thing, aliases []string) string {

	if len(aliases) == 0 {
		return "You examine this and that and find nothing special. Maybe if you examined something specific?"
	}

	what, _ := WhatWhere(aliases[0], t)

	if what == nil {
		return "You see no '" + aliases[0] + "' to examine."
	}

	if veto := CheckVetoes("EXAMINE", what); veto != nil {
		return veto.Message()
	}

	buff := make([]byte, 0, 1024)

	if n := attr.Name().Find(what); n != nil {
		buff = append(buff, "You examine "...)
		buff = append(buff, n.Name()...)
		buff = append(buff, "."...)
	}

	for _, d := range attr.Description().FindAll(what) {
		buff = append(buff, " "...)
		buff = append(buff, d.Description()...)
	}

	if i := attr.Inventory().Find(what); i != nil {
		buff = append(buff, " "...)
		buff = append(buff, i.Contents()...)
	}

	return string(buff)
}
