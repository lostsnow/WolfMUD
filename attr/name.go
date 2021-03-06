// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/event"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Name attribute.
func init() {
	internal.AddMarshaler((*Name)(nil), "name")
}

// Initialise event finder
func init() {
	event.FindName = FindName
}

// Name implements an attribute for giving a Thing a short name. The name
// should generally start with 'a', 'an' or 'some' except for proper names. For
// example: an apple, some apples, an orange, Diddymus.
type Name struct {
	Attribute
	name string
}

// Some interfaces we want to make sure we implement
var (
	_ has.Name = &Name{}
)

// Name implements an attribute for naming Things. It is used when referring to
// or listing Things. For example if there is a sword it could have the name of
// 'a sword'. Then manipulating it you could see the following messages:
//
//	You see a sword here.
//	You pick up a sword.
//	You examine a sword.
//	You start to wield a sword.
//
// Messages such as the examples would typically be general messages with a
// placeholder for the name of the Thing. For example:
//
//	You see %s here.
//	You pick up %s.
//	You examine %s.
//	You start to wield %s.
//
// It is therefore important to take this into consideration when choosing
// names for Things.
func NewName(n string) *Name {
	return &Name{Attribute{}, n}
}

// FindName searches the attributes of the specified Thing for attributes that
// implement has.Name returning the first match it finds or a *Name typed nil
// otherwise.
func FindName(t has.Thing) has.Name {
	return t.FindAttr((*Name)(nil)).(has.Name)
}

// Is returns true if passed attribute implements a name else false.
func (*Name) Is(a has.Attribute) bool {
	_, ok := a.(has.Name)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (n *Name) Found() bool {
	return n != nil
}

// Unmarshal is used to turn the passed data into a new Name attribute.
func (*Name) Unmarshal(data []byte) has.Attribute {
	return NewName(decode.String(data))
}

// Marshal returns a tag and []byte that represents the receiver.
func (n *Name) Marshal() (tag string, data []byte) {
	return "name", encode.String(n.name)
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (n *Name) Dump(node *tree.Node) *tree.Node {
	return node.Append("%p %[1]T - %q", n, n.name)
}

// Name returns the name stored in the attribute. If the receiver is nil or the
// name is an empty string the specified preset will be returned instead. This
// allows for a generic preset name such as someone, something or somewhere to
// be returned for things without names.
func (n *Name) Name(preset string) string {
	switch {
	case n == nil:
		return preset
	case n.name == "":
		return preset
	default:
		return n.name
	}
}

// repl holds from/to replacement strings used by TheName.
var repl = map[string]string{
	"a":    "the",
	"A":    "The",
	"an":   "the",
	"An":   "The",
	"AN":   "THE",
	"some": "the",
	"Some": "The",
	"SOME": "THE",
}

// TheName returns the name stored in the attribute, as per Name, but with the
// leading "A ", "An " or "Some " changed to "The ". The case of the 't' in
// 'the' is the same as the case of word replaced. For example:
//
//	      A frog -> The frog
//	      a frog -> the frog
//	    An apple -> The apple
//	    an apple -> the apple
//	 Some apples -> The apples
//	 some apples -> the apples
//
func (n *Name) TheName(preset string) string {

	name := n.Name(preset)

	if len(name) == 0 {
		return name
	}

	// Bail out quickly if initial letter doesn't match possible choices
	if name[0] != 'A' && name[0] != 'a' && name[0] != 'S' && name[0] != 's' {
		return name
	}

	for from := range repl {
		l := len(from)
		if l >= len(name) || name[l] != ' ' {
			continue
		}
		if to, ok := repl[name[0:l]]; ok {
			return to + name[l:]
		}
	}
	return name
}

// Copy returns a copy of the Name receiver.
func (n *Name) Copy() has.Attribute {
	if n == nil {
		return (*Name)(nil)
	}
	return NewName(n.name)
}
