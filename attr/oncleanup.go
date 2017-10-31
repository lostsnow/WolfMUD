// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
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

// Found returns false if the receiver is nil otherwise true.
func (oc *OnCleanup) Found() bool {
	return oc != nil
}

// Unmarshal is used to turn the passed data into a new OnCleanup attribute.
func (*OnCleanup) Unmarshal(data []byte) has.Attribute {
	return NewOnCleanup(decode.String(data))
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
