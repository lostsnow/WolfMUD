// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"log"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Wielding attribute.
func init() {
	internal.AddMarshaler((*Wielding)(nil), "wielding")
}

// Wielding implements an attribute representing items that are being wielded
// by a Thing.
//
// BUG(diddymus): At the moment the Wielding attribute only represents the
// initial state of items being wielded when a Thing is unmarshaled - it is not
// dynamically updated as items are wielded and removed.
type Wielding struct {
	Attribute
	refs []string
}

// NewWielding returns a Wielding attribute initialised with the passed
// references. The references should be those returned by Thing.Ref(), which
// are normally the content of the Ref field from the record jar the Thing was
// loaded from.
func NewWielding(ref ...string) *Wielding {
	return &Wielding{Attribute{}, ref}
}

// FindWielding searches the attributes of the specified Thing for attributes
// that implement has.Wielding returning the first match it finds or a
// *Wielding typed nil otherwise.
func FindWielding(t has.Thing) has.Wielding {
	return t.FindAttr((*Wielding)(nil)).(has.Wielding)
}

// Is returns true if passed attribute implements wielding else false.
func (*Wielding) Is(a has.Attribute) bool {
	_, ok := a.(has.Wielding)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (w *Wielding) Found() bool {
	return w != nil
}

// Unmarshal is used to turn the passed data into a new Wielding attribute.
func (*Wielding) Unmarshal(data []byte) has.Attribute {
	return NewWielding(decode.KeywordList(data)...)
}

// load post-unmarshal hook to actually cause items to be wielded when a Thing
// is loaded and unmarshaled.
func (w *Wielding) load() {
	p := w.Parent()
	b := FindBody(p)
	if !b.Found() {
		log.Printf("  %q, ref: %q - body not found, cannot wield items",
			FindName(p).Name("Someone"), p.Ref())
		return
	}

	var t has.Thing
	var W has.Wieldable

	i := FindInventory(p)
	for _, ref := range w.refs {
		if t = i.SearchByRef(ref); t == nil {
			log.Printf("  %q, ref: %q - ref: %q, wieldable item not found",
				FindName(p).Name("Someone"), p.Ref(), ref)
			continue
		}
		if W = FindWieldable(t); !W.Found() {
			log.Printf("  %q, ref: %q - %q, ref: %q, not a wieldable item",
				FindName(p).Name("Someone"), p.Ref(), FindName(t).Name("Something"), ref)
			continue
		}
		b.Wield(W)
	}
}

// resetHook to cause items to be wielded when a Thing is reset.
func (w *Wielding) resetHook() {

	var (
		t has.Thing
		b has.Body
		i has.Inventory
	)

	p := w.Parent()
	if b = FindBody(p); !b.Found() {
		return
	}
	if i = FindInventory(p); !i.Found() {
		return
	}

	for _, ref := range w.refs {
		if t = i.SearchByRef(ref); t != nil {
			b.Wield(FindWieldable(t))
		}
	}
}

// Marshal returns a tag and []byte that represents the receiver.
func (w *Wielding) Marshal() (tag string, data []byte) {
	refs := []string{}
	for _, t := range FindBody(w.Parent()).Wielding() {
		refs = append(refs, t.UID())
	}
	return "wielding", encode.KeywordList(refs)
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (w *Wielding) Dump(node *tree.Node) *tree.Node {
	return node.Append("%p %[1]T - %q", w, w.refs)
}

// Wielded returns a string slice of Thing references for items that are being
// wielded.
func (w *Wielding) Wielded() []string {
	return w.refs
}

// Copy returns a copy of the Wielding receiver.
func (w *Wielding) Copy() has.Attribute {
	if w == nil {
		return (*Wielding)(nil)
	}
	return NewWielding(w.refs...)
}
