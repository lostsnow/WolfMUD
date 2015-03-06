// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type Writing struct {
	Attribute
	writing string
}

// Some interfaces we want to make sure we implement
var (
	_ has.Description = &Writing{}
	_ has.Writing     = &Writing{}
)

func NewWriting(w string) *Writing {
	return &Writing{Attribute{}, w}
}

func FindWriting(t has.Thing) has.Writing {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Writing); ok {
			return a
		}
	}
	return nil
}

func (w *Writing) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", w, w.writing)}
}

func (w *Writing) Writing() string {
	return w.writing
}

func (w *Writing) Description() string {
	return "It has something written on it."
}
