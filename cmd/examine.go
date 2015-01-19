// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
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

	name := ""
	description := ""
	contents := ""

	for _, a := range what.Attrs() {
		switch a := a.(type) {
		case has.Name:
			name = a.Name()
		case has.Description:
			description += " " + a.Description()
		case has.Inventory:
			contents = " " + a.Contents()
		}
	}

	return "You examine " + name + "." + description + contents
}
