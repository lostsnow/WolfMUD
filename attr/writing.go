// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD-mini.git/has"
)

type writing struct {
	attribute
	writing string
}

// Some interfaces we want to make sure we implement
var _ has.Attribute = &writing{}
var _ has.Description = &writing{}
var _ has.Writing = &writing{}

func NewWriting(w string) *writing {
	return &writing{attribute{}, w}
}

func FindWriting(t has.Thing) (w has.Writing) {
	w, _ = t.Find(&w).(has.Writing)
	return
}

func (w *writing) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", w, w.writing)}
}

func (w *writing) Writing() string {
	return w.writing
}

func (w *writing) Description() string {
	return "It has something written on it."
}
