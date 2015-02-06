// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"

	"strings"
)

func Dump(t has.Thing, aliases []string) (msg string, ok bool) {

	if len(aliases) == 0 {
		msg = "What do you want to dump?"
		return
	}

	what, _ := whatWhere(aliases[0], t)

	// As a last resort instead of looking 'IN where we are' look 'AT where we
	// are' - whatWhere does not check if the what is also the where.
	if what == nil {
		if l := attr.Locate().Find(t); l != nil {
			if where := l.Where(); where != nil {
				if a := attr.Alias().Find(where); a != nil {
					if a.HasAlias(aliases[0]) {
						what = where
					}
				}
			}
		}
	}

	if what == nil {
		msg = "Nothing with alias '" + aliases[0] + "' found to dump."
		return
	}

	msg = strings.Join(what.Dump(), "\n")
	return msg, true
}
