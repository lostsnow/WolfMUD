// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"
)

// Register marshaler for OnReset attribute.
func init() {
	internal.AddMarshaler((*OnReset)(nil), "OnReset")
}

// OnReset implements an attribute to provide a reset or respawn message for a
// Thing.
type OnReset struct {
	Attribute
	text string
}

// Some interfaces we want to make sure we implement
var (
	_ has.OnReset = &OnReset{}
)

// NewOnReset returns a new OnReset attribute initialised with the specified
// message.
func NewOnReset(text string) *OnReset {
	return &OnReset{Attribute{}, text}
}

// FindOnReset searches the attributes of the specified Thing for attributes
// that implement has.OnReset returning the first match it finds or a *OnReset
// typed nil otherwise.
func FindOnReset(t has.Thing) has.OnReset {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.OnReset); ok {
			return a
		}
	}
	return (*OnReset)(nil)
}

// Found returns false if the receiver is nil otherwise true.
func (or *OnReset) Found() bool {
	return or != nil
}

// Unmarshal is used to turn the passed data into a new OnReset attribute.
func (*OnReset) Unmarshal(data []byte) has.Attribute {
	return NewOnReset(recordjar.Decode.String(data))
}

func (or *OnReset) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", or, or.text)}
}

// ResetText returns the reset or respawn message to be used for a Thing.
func (or *OnReset) ResetText() string {
	if or == nil {
		return ""
	}
	return or.text
}

// Copy returns a copy of the OnReset receiver.
func (or *OnReset) Copy() has.Attribute {
	if or == nil {
		return (*OnReset)(nil)
	}
	return NewOnReset(or.text)
}
