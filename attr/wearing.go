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

// Register marshaler for Wearing attribute.
func init() {
	internal.AddMarshaler((*Wearing)(nil), "wearing")
}

// Wearing implements an attribute representing items that are being worn by a
// Thing.
//
// BUG(diddymus): At the moment the Wearing attribute only represents the
// initial state of items being worn when a Thing is unmarshaled - it is not
// dynamically updated as items are worn and removed.
type Wearing struct {
	Attribute
	refs []string
}

// NewWearing returns a Wearing attribute initialised with the passed
// references. The references should be those returned by Thing.Ref(), which
// are normally the content of the Ref field from the record jar the Thing was
// loaded from.
func NewWearing(ref ...string) *Wearing {
	return &Wearing{Attribute{}, ref}
}

// FindWearing searches the attributes of the specified Thing for attributes
// that implement has.Wearing returning the first match it finds or a *Wearing
// typed nil otherwise.
func FindWearing(t has.Thing) has.Wearing {
	return t.FindAttr((*Wearing)(nil)).(has.Wearing)
}

// Is returns true if passed attribute implements wearing else false.
func (*Wearing) Is(a has.Attribute) bool {
	_, ok := a.(has.Wearing)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (w *Wearing) Found() bool {
	return w != nil
}

// Unmarshal is used to turn the passed data into a new Wearing attribute.
func (*Wearing) Unmarshal(data []byte) has.Attribute {
	return NewWearing(decode.KeywordList(data)...)
}

// load post-unmarshal hook to actually cause items to be worn when a Thing is
// loaded and unmarshaled.
func (w *Wearing) load() {
	p := w.Parent()
	b := FindBody(p)
	if !b.Found() {
		log.Printf("  %q, ref: %q - body not found, cannot wear items",
			FindName(p).Name("Someone"), p.Ref())
		return
	}

	var t has.Thing
	var W has.Wearable

	i := FindInventory(p)
	for _, ref := range w.refs {
		if t = i.SearchByRef(ref); t == nil {
			log.Printf("  %q, ref: %q - ref: %q, wearable item not found",
				FindName(p).Name("Someone"), p.Ref(), ref)
			continue
		}
		if W = FindWearable(t); !W.Found() {
			log.Printf("  %q, ref: %q - %q, ref: %q, not a wearable item",
				FindName(p).Name("Someone"), p.Ref(), FindName(t).Name("Something"), ref)
			continue
		}
		b.Wear(W)
	}
}

// resetHook to cause items to be worn when a Thing is reset.
func (w *Wearing) resetHook() {

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
			b.Wear(FindWearable(t))
		}
	}
}

// Marshal returns a tag and []byte that represents the receiver.
func (w *Wearing) Marshal() (tag string, data []byte) {
	refs := []string{}
	for _, t := range FindBody(w.Parent()).Wearing() {
		refs = append(refs, t.UID())
	}
	return "wearing", encode.KeywordList(refs)
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (w *Wearing) Dump(node *tree.Node) *tree.Node {
	return node.Append("%p %[1]T - %q", w, w.refs)
}

// Worn returns a string slice of Thing references for items that are being
// worn.
func (w *Wearing) Worn() []string {
	return w.refs
}

// Copy returns a copy of the Wearing receiver.
func (w *Wearing) Copy() has.Attribute {
	if w == nil {
		return (*Wearing)(nil)
	}
	return NewWearing(w.refs...)
}
