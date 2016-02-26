// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/has"
)

// Writing implements an attribute that allows for writing to be put onto any
// Thing so that it can be read.
//
// TODO: Writing currently assumes the text is written onto a Thing. However it
// could also be carved, burnt, painted, etc. onto a Thing. It also assumes the
// text is in a common language known to all. If language were implemented we
// could write in common, elvish, dwarfish, ancient runes, secret code or
// anything else with the text only being readable by those who know the
// relevant language. See also the Writing Description method.
type Writing struct {
	Attribute
	writing string
}

// Some interfaces we want to make sure we implement
var (
	_ has.Description = &Writing{}
	_ has.Writing     = &Writing{}
)

// NewWriting returns a new Writing attribute initialised with the specified
// writing/text.
func NewWriting(w string) *Writing {
	return &Writing{Attribute{}, w}
}

// FindWriting searches the attributes of the specified Thing for attributes
// that implement has.Writing returning the first match it finds or a *Writing
// typed nil otherwise.
func FindWriting(t has.Thing) has.Writing {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Writing); ok {
			return a
		}
	}
	return (*Writing)(nil)
}

// Found returns false if the receiver is nil otherwise true.
func (w *Writing) Found() bool {
	return w != nil
}

func (w *Writing) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", w, w.writing)}
}

// Writing returns the text that has been written.
func (w *Writing) Writing() (writing string) {
	if w != nil {
		writing = w.writing
	}
	return
}

// Description automatically adds the specified text to the description of a
// Thing that has a has.Writing attribute.
//
// FIXME: This should return a message based on the type of writing: runes,
// painting, carving etc. See also TODO for the Writing type.
func (w *Writing) Description() string {
	return "It has something written on it."
}
