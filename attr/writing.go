// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Writing attribute.
func init() {
	internal.AddMarshaler((*Writing)(nil), "writing")
}

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
	return t.FindAttr((*Writing)(nil)).(has.Writing)
}

// Is returns true if passed attribute implements writing else false.
func (*Writing) Is(a has.Attribute) bool {
	_, ok := a.(has.Writing)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (w *Writing) Found() bool {
	return w != nil
}

// Unmarshal is used to turn the passed data into a new Writing attribute.
func (*Writing) Unmarshal(data []byte) has.Attribute {
	return NewWriting(decode.String(data))
}

// Marshal returns a tag and []byte that represents the receiver.
func (w *Writing) Marshal() (tag string, data []byte) {
	return "writing", encode.String(w.writing)
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (w *Writing) Dump(node *tree.Node) *tree.Node {
	return node.Append("%p %[1]T - %q", w, w.writing)
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

// Copy returns a copy of the Writing receiver.
func (w *Writing) Copy() has.Attribute {
	if w == nil {
		return (*Writing)(nil)
	}
	return NewWriting(w.writing)
}
