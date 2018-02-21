// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"math/rand"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
)

// Register marshaler for OnAction attribute.
func init() {
	internal.AddMarshaler((*OnAction)(nil), "OnAction")
}

// OnAction implements an attribute to provide action messages for a Thing.
type OnAction struct {
	Attribute
	actions []string
}

// Some interfaces we want to make sure we implement
var (
	_ has.OnAction = &OnAction{}
)

// NewOnAction returns a new OnAction attribute initialised with the
// specified messages.
func NewOnAction(actions []string) *OnAction {
	return &OnAction{Attribute{}, actions}
}

// FindOnAction searches the attributes of the specified Thing for attributes
// that implement has.OnAction returning the first match it finds or a
// *OnAction typed nil otherwise.
func FindOnAction(t has.Thing) has.OnAction {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.OnAction); ok {
			return a
		}
	}
	return (*OnAction)(nil)
}

// Found returns false if the receiver is nil otherwise true.
func (oa *OnAction) Found() bool {
	return oa != nil
}

// Unmarshal is used to turn the passed data into a new OnAction attribute.
func (*OnAction) Unmarshal(data []byte) has.Attribute {
	return NewOnAction(decode.StringList(data))
}

// Marshal returns a tag and []byte that represents the receiver.
func (oa *OnAction) Marshal() (tag string, data []byte) {
	return "onaction", encode.StringList(oa.actions)
}

func (oa *OnAction) Dump() (buf []string) {
	buf = append(buf, DumpFmt("%p %[1]T %d actions:", oa, len(oa.actions)))
	for i, action := range oa.actions {
		buf = append(buf, DumpFmt("  #%d: %q", i, action))
	}
	return
}

// ActionText returns a random action message for a Thing. The message is
// chosen from the list of messages available.
func (oa *OnAction) ActionText() string {
	if oa == nil {
		return ""
	}
	i := rand.Intn(len(oa.actions))
	return oa.actions[i]
}

// Copy returns a copy of the OnAction receiver.
func (oa *OnAction) Copy() has.Attribute {
	if oa == nil {
		return (*OnAction)(nil)
	}
	return NewOnAction(oa.actions)
}
