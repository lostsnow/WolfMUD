// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
)

// Register marshaler for OnCleanup attribute.
func init() {
	internal.AddMarshaler((*OnCleanup)(nil), "OnCleanup")
}

// OnCleanup implements an attribute to provide a clean up message for a
// Thing.
type OnCleanup struct {
	Attribute
	text string
}

// Some interfaces we want to make sure we implement
var (
	_ has.OnCleanup = &OnCleanup{}
)

// NewOnCleanup returns a new OnCleanup attribute initialised with the
// specified message.
func NewOnCleanup(text string) *OnCleanup {
	return &OnCleanup{Attribute{}, text}
}

// FindOnCleanup searches the attributes of the specified Thing for attributes
// that implement has.OnCleanup returning the first match it finds or a
// *OnCleanup typed nil otherwise.
func FindOnCleanup(t has.Thing) has.OnCleanup {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.OnCleanup); ok {
			return a
		}
	}
	return (*OnCleanup)(nil)
}

// Is returns true if passed attribute implements an 'on cleanup' else false.
func (*OnCleanup) Is(a has.Attribute) bool {
	_, ok := a.(has.OnCleanup)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (oc *OnCleanup) Found() bool {
	return oc != nil
}

// Unmarshal is used to turn the passed data into a new OnCleanup attribute.
func (*OnCleanup) Unmarshal(data []byte) has.Attribute {
	return NewOnCleanup(decode.String(data))
}

// Marshal returns a tag and []byte that represents the receiver.
func (oc *OnCleanup) Marshal() (tag string, data []byte) {
	return "oncleanup", encode.String(oc.text)
}

func (oc *OnCleanup) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", oc, oc.text)}
}

// CleanupText returns the clean up message to be used for a Thing.
func (oc *OnCleanup) CleanupText() string {
	if oc == nil {
		return ""
	}
	return oc.text
}

// Copy returns a copy of the OnCleanup receiver.
func (oc *OnCleanup) Copy() has.Attribute {
	if oc == nil {
		return (*OnCleanup)(nil)
	}
	return NewOnCleanup(oc.text)
}
