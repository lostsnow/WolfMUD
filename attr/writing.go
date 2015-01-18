// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type writing struct {
	parent
	writing string
}

// Some interfaces we want to make sure we implement
var _ has.Attribute = &writing{}
var _ has.Writing = &writing{}

func NewWriting(w string) *writing {
	return &writing{parent{}, w}
}

func FindWriting(t has.Thing) has.Writing {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Writing); ok {
			return a
		}
	}
	return nil
}

func (w *writing) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", w, w.writing)}
}

func (w *writing) Writing() string {
	return w.writing
}
