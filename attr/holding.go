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

// Register marshaler for Holding attribute.
func init() {
	internal.AddMarshaler((*Holding)(nil), "holding")
}

// Holding implements an attribute representing items that are being held by a
// Thing.
//
// BUG(diddymus): At the moment the Holding attribute only represents the
// initial state of items being held when a Thing is unmarshaled - it is not
// dynamically updated as items are held and removed.
type Holding struct {
	Attribute
	refs []string
}

// NewHolding returns a Holding attribute initialised with the passed
// references. The references should be those returned by Thing.Ref(), which
// are normally the content of the Ref field from the record jar the Thing was
// loaded from.
func NewHolding(ref ...string) *Holding {
	return &Holding{Attribute{}, ref}
}

// FindHolding searches the attributes of the specified Thing for attributes
// that implement has.Holding returning the first match it finds or a *Holding
// typed nil otherwise.
func FindHolding(t has.Thing) has.Holding {
	return t.FindAttr((*Holding)(nil)).(has.Holding)
}

// Is returns true if passed attribute implements holding else false.
func (*Holding) Is(a has.Attribute) bool {
	_, ok := a.(has.Holding)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (h *Holding) Found() bool {
	return h != nil
}

// Unmarshal is used to turn the passed data into a new Holding attribute.
func (*Holding) Unmarshal(data []byte) has.Attribute {
	return NewHolding(decode.KeywordList(data)...)
}

// loadHook to hold items specified by Holding attribute once a Thing has been
// unmarshaled and we have access to Inventory content.
func (h *Holding) loadHook() {

	var (
		t    has.Thing
		b    has.Body
		what has.Holdable
	)

	p := h.Parent()
	pName := FindName(p).Name("Something")

	if b = FindBody(p); !b.Found() {
		log.Printf("  %q, ref: %q - body not found, cannot hold items", pName, p.Ref())
		return
	}

	i := FindInventory(p)
	for _, ref := range h.refs {
		if t = i.SearchByRef(ref); t == nil {
			log.Printf("  %q, ref: %q, holdable item not found in inventory, ref: %q", pName, p.Ref(), ref)
			continue
		}
		if what = FindHoldable(t); !what.Found() {
			log.Printf("  %q, ref: %q - %q, not a holdable item, ref: %q", pName, p.Ref(), FindName(t).Name("Something"), ref)
			continue
		}
		b.Hold(what)
	}
}

// resetHook to cause items to be held when a Thing is reset.
func (h *Holding) resetHook() {

	var (
		t has.Thing
		b has.Body
		i has.Inventory
	)

	p := h.Parent()
	if b = FindBody(p); !b.Found() {
		return
	}
	if i = FindInventory(p); !i.Found() {
		return
	}

	for _, ref := range h.refs {
		if t = i.SearchByRef(ref); t != nil {
			b.Hold(FindHoldable(t))
		}
	}
}

// Marshal returns a tag and []byte that represents the receiver.
func (h *Holding) Marshal() (tag string, data []byte) {
	refs := []string{}
	for _, t := range FindBody(h.Parent()).Holding() {
		refs = append(refs, t.UID())
	}
	return "holding", encode.KeywordList(refs)
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (h *Holding) Dump(node *tree.Node) *tree.Node {
	return node.Append("%p %[1]T - %q", h, h.refs)
}

// Held returns a string slice of Thing references for items that are being
// held.
func (h *Holding) Held() []string {
	return h.refs
}

// Copy returns a copy of the Holding receiver.
func (h *Holding) Copy() has.Attribute {
	if h == nil {
		return (*Holding)(nil)
	}
	return NewHolding(h.refs...)
}
