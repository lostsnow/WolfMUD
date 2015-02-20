// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

func CheckVetoes(cmd string, what has.Thing) has.Veto {
	if vetoes := attr.Vetoes().Find(what); vetoes != nil {
		if veto := vetoes.Check(cmd); veto != nil {
			return veto
		}
	}

	return nil
}
