// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/text"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Thing is a container for Attributes. Everything in WolfMUD is constructed by
// creating a Thing and then adding Attributes to it which implement specific
// functionality. Concurrent access to a Thing is safe.
type Thing struct {
	uid string

	rwmutex sync.RWMutex
	attrs   []has.Attribute
}

// Some interfaces we want to make sure we implement
var (
	_ has.Thing = &Thing{}
)

// ThingCount is a channel storing the current number of Things. The value
// should only be incremented by NewThing and decremented by Free. ThingCount
// is a channel so that the value can be updated concurrently and shared with
// e.g. the stats package.
var ThingCount chan uint64

// init sets up and initialises ThingCount.
func init() {
	ThingCount = make(chan uint64, 1)
	ThingCount <- 0
}

// NewThing returns a new Thing initialised with the specified Attributes.
// Attributes can also be dynamically modified using Add and Remove methods.
// If Debug.Things is true a message will be written to the log indicating a
// new Thing has been created. A finalizer will also be registered to write a
// message when the thing is garbage collected.
func NewThing(a ...has.Attribute) *Thing {
	t := &Thing{uid: <-internal.NextUID, attrs: []has.Attribute{}}

	t.Add(a...)

	ThingCount <- <-ThingCount + 1

	if config.Debug.Things {
		runtime.SetFinalizer(t, func(t has.Thing) {
			log.Printf("Finalizing: %s", t)
		})
		log.Printf("NewThing: %s: %q", t, FindName(t).Name("?"))
	}

	return t
}

// Free is used to clean-up/release references to all Attribute for a Thing.
// When a Thing is finished with calling Free helps the garbage collector to
// reclaim objects. It can also help to break cyclic references that could
// prevent garbage collection.
func (t *Thing) Free() {
	if t == nil {
		return
	}

	t.rwmutex.Lock()

	if t.uid != "" && t.attrs == nil {
		log.Printf("Warning, already freed: %s", t)
	}

	for i := range t.attrs {
		t.attrs[i].Free()
		t.attrs[i] = nil
	}
	t.attrs = nil

	if t.uid != "" {
		ThingCount <- <-ThingCount - 1
	}

	t.rwmutex.Unlock()
}

// Freed returns true if Free has been called on the Thing, else false.
func (t *Thing) Freed() (b bool) {
	t.rwmutex.RLock()
	b = t.attrs == nil
	t.rwmutex.RUnlock()
	return
}

// Add is used to add the passed Attributes to a Thing. When an Attribute is
// added its parent is set to reference the Thing it was added to. This allows
// an Attribute to find and query the parent Thing about other Attributes the
// Thing may have.
func (t *Thing) Add(a ...has.Attribute) {
	t.rwmutex.Lock()
	for _, a := range a {
		a.SetParent(t)
		t.attrs = append(t.attrs, a)
	}
	t.rwmutex.Unlock()
}

// Remove is used to remove the passed Attributes from a Thing. There is no
// indication if an Attribute cannot actually be removed. When an Attribute is
// removed its parent is set to nil. When an Attribute is removed and is no
// longer required the Attribute's Free method should be called.
func (t *Thing) Remove(a ...has.Attribute) {
	t.rwmutex.Lock()
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(t.attrs) - 1; j >= 0; j-- {
			if a[i] == t.attrs[j] {
				t.attrs[j].SetParent(nil)
				copy(t.attrs[j:], t.attrs[j+1:])
				t.attrs[len(t.attrs)-1] = nil
				t.attrs = t.attrs[:len(t.attrs)-1]
				break
			}
		}
	}
	t.rwmutex.Unlock()
}

// FindAttr searches the attributes of the Thing for attributes that implement
// the passed Attribute cmp returning the first match it finds or cmp
// otherwise. The comparison is performed by calling cmp.Is on the attributes
// of the Thing. It is usual for cmp to be a typed nil attribute and for the
// returned attribute to be converted to the general has interface type for
// cmp. For an example see the attr.FindName function.
func (t *Thing) FindAttr(cmp has.Attribute) has.Attribute {
	t.rwmutex.RLock()
	for _, a := range t.attrs {
		if cmp.Is(a) {
			t.rwmutex.RUnlock()
			return a
		}
	}
	t.rwmutex.RUnlock()
	return cmp
}

// FindAttrs searches the attributes of the Thing for attributes that implement
// the passed Attribute cmp returning a slice of all the matches it finds or a
// nil slice otherwise. The comparison is performed by calling cmp.Is on the
// attributes of the Thing. It is usual for cmp to be a typed nil attribute and
// for the returned attribute to be converted to the general has interface
// type for cmp. For an example see the attr.FindAllDescription function.
func (t *Thing) FindAttrs(cmp has.Attribute) (attrs []has.Attribute) {
	t.rwmutex.RLock()
	for _, a := range t.attrs {
		if cmp.Is(a) {
			attrs = append(attrs, a)
		}
	}
	t.rwmutex.RUnlock()
	return
}

// ignoredFields is a list of known field names that should be ignored by
// Unmarshal as there is no corresponding Attribute to unmarshal the field's
// data. If the field isn't ignored we just get extra warnings in the log when
// unmarshaling is attempted.
var ignoredFields = text.Dictionary("ref", "location", "zonelinks")

// Unmarshal unmarshals a Thing from a recordjar record containing all of the
// Attribute to be added. The recno is the record number in the recordjar for
// this record. It is passed so that we can give informative messages if errors
// are found. If the record number is not known -1 should be passed instead.
func (t *Thing) Unmarshal(recno int, record recordjar.Record) {

	var (
		m  has.Marshaler
		a  has.Attribute
		ok bool
	)

	// Go through the fields in the record
	for field, data := range record {

		// Some known fields without attributes or marshalers we don't want to
		// try and unmarshal so we ignore.
		if ignoredFields.Contains(field) {
			continue
		}

		// Look for a marshaler for the field name
		if m, ok = internal.Marshalers[field]; !ok {
			if recno == -1 {
				log.Printf("Unknown attribute: %s", field)
			} else {
				log.Printf("[Record: %d] Unknown attribute: %s", recno, field)
			}
			continue
		}

		// Unmarshal the data into an Attribute and add it to the Thing as long as
		// the returned, unmarshaled Attribute is not an untyped nil.
		if a = m.Unmarshal(data); a != nil {
			t.Add(a)
		}
	}

	return
}

// Marshal marshals a Thing to a recordjar record containing all of the
// Attribute details.
func (t *Thing) Marshal() recordjar.Record {

	var (
		tag  string
		data []byte
	)

	t.rwmutex.RLock()
	rec := recordjar.Record{}
	rec["ref"] = []byte(t.UID())
	for _, a := range t.attrs {
		tag, data = a.Marshal()
		if tag == "" {
			continue
		}
		rec[tag] = data
	}
	t.rwmutex.RUnlock()
	return rec
}

// DumpToLog is a convenience method for dumping the current state of a Thing
// to the log. The information is annotated with the file and line number where
// the dump was taken and specified label.
//
// NOTE: Care should be take if debugging Thing itself (Add, Remove, Free) as
// Dump will acquire a read lock on the rwmutex.
func (t *Thing) DumpToLog(label string) {
	_, file, line, _ := runtime.Caller(1)
	dbg := tree.Tree{}
	dbg.Indent, dbg.Width, dbg.Offset = 20, 110, 13
	t.Dump(dbg.Branch())
	log.Printf("Debug @ %s:%d (%q)\n"+dbg.Render(), file, line, label)
}

// Dump adds Thing information to the passed tree.Node for debugging and
// returns the new node. A new branch is created on the node which is passed to
// each of the Thing's attributes to add their information. This may continue
// recursivly as in the case of containers.
//
// NOTE: Care should be take if debugging Thing itself (Add, Remove, Free) as
// Dump will acquire a read lock on the rwmutex.
func (t *Thing) Dump(node *tree.Node) *tree.Node {
	t.rwmutex.RLock()

	// Manually try to find the Name of the Thing as FindName could panic
	name := "???"
	if t.attrs != nil {
		for _, a := range t.attrs {
			if a, ok := a.(has.Name); ok {
				name = a.Name(name)
			}
		}
	}

	node = node.Append("%s (%q), collectable: %t, attributes: %d",
		t, name, t.Collectable(), len(t.attrs),
	)

	branch := node.Branch()
	for _, a := range t.attrs {
		dump(branch, a)
	}

	t.rwmutex.RUnlock()
	return node
}

// dump will attempt to add the specified Attribute information to the passed
// node. If the attempt causes a panic the error will be appended to the node
// instead of the Attribute information.
func dump(node *tree.Node, a has.Attribute) {
	defer func() {
		if r := recover(); r != nil {
			node.Append("%p %[1]T - (error %v)", a, r)
			return
		}
	}()
	a.Dump(node)
}

// Copy returns a copy of the Thing receiver. Associated Attributes will be
// copied. However, the copy is not recursive and will not copy the content of
// Inventory. To make a copy that includes Inventory content the DeepCopy
// method should be used instead.
//
// BUG(diddymus): This method specifically checks for a *attr.Inventory, which
// is currently the only implementation of has.Inventory - if this changes this
// method will need updating.
func (t *Thing) Copy() has.Thing {
	if t == nil {
		return (*Thing)(nil)
	}
	t.rwmutex.RLock()
	na := make([]has.Attribute, len(t.attrs), len(t.attrs))
	for i, a := range t.attrs {
		// If we find an inventory provide a new one, don't copy - this prevents
		// recursive copying
		if _, ok := a.(*Inventory); ok {
			na[i] = NewInventory()
		} else {
			na[i] = a.Copy()
		}
	}
	t.rwmutex.RUnlock()
	return NewThing(na...)
}

// DeepCopy returns a copy of the Thing receiver recursing into Inventory. The
// copy will be made recursively copying all associated Attribute and Thing. To
// make a non-recursive copy (excluding Inventory content) the Copy method
// should be used instead.
func (t *Thing) DeepCopy() has.Thing {
	if t == nil {
		return (*Thing)(nil)
	}
	t.rwmutex.RLock()
	na := make([]has.Attribute, len(t.attrs), len(t.attrs))
	for i, a := range t.attrs {
		na[i] = a.Copy()
	}
	t.rwmutex.RUnlock()
	return NewThing(na...)
}

// SetOrigins updates the origin for the Thing to its containing Inventory and
// recursivly sets the origins for the content of a Thing's Inventory if it has
// one.
func (t *Thing) SetOrigins() {
	if t == nil {
		return
	}

	// Set our origin to that of the parent Inventory
	if l := FindLocate(t); l.Found() {
		l.SetOrigin(l.Where())
	}

	// Find our Inventory
	i := FindInventory(t)
	if !i.Found() {
		return
	}

	// Set the origin for everything in our Inventory including disabled items
	for _, t := range i.Everything() {
		t.SetOrigins()
	}
	for _, t := range i.Disabled() {
		t.SetOrigins()
	}
}

// ClearOrigins sets the origin for the Thing to nil and recursivly sets the
// origins to nil for the content of a Thing's Inventory if it has one.
func (t *Thing) ClearOrigins() {
	if t == nil {
		return
	}

	// Clear our origin
	if l := FindLocate(t); l.Found() {
		l.SetOrigin(nil)
	}

	// Find our Inventory
	i := FindInventory(t)
	if !i.Found() {
		return
	}

	// Clear the origin for items in our Inventory including disabled items
	for _, t := range i.Everything() {
		t.ClearOrigins()
	}
	for _, t := range i.Disabled() {
		t.ClearOrigins()
	}
}

// Collectable returns true if a Thing can be kept by a player, otherwise
// returns false. This is a helper routine so that the definition of what is
// considered collectable can be easily changed.
func (t *Thing) Collectable() bool {
	o := FindLocate(t).Origin()

	// Collectable if there is no origin...
	collectable := o == nil || !o.Found()

	// ...unless they are a Player
	if collectable && FindPlayer(t).Found() {
		collectable = false
	}

	return collectable
}

// UID returns the unique identifier for a specific Thing or an empty string if
// the unique ID is unavailable. The unique ID should be automatically assigned
// to any Thing created by calling NewThing or Copy.
func (t *Thing) UID() string {
	if t == nil {
		return ""
	}
	return t.uid
}

// NotUnique marks a Thing as no longer being unique and clears the Thing's
// UID. It also decrements ThingCount by one to account for the fact that
// calling Free will no longer decrement ThingCount for multiple references of
// this Thing. Calling NotUnique on a Thing more than once is safe.
//
// You almost never, ever, want to call this function! The only time this
// should be used is when creating temporary stores - such as when loading
// zones or players.
func (t *Thing) NotUnique() {
	if t == nil || t.uid == "" {
		return
	}

	ThingCount <- <-ThingCount - 1
	t.uid = ""
}

// String causes a Thing to implement the Stringer interface so that a Thing
// can print information about itself. The format of the string is:
//
//  <address> <type> - uid: <unique ID>
//
//  0xc420108630 *attr.Thing - uid: #UID-6M
//
func (t *Thing) String() string {
	return fmt.Sprintf("%p %[1]T - uid: %s", t, t.UID())
}

// Things is a type of slice *Thing. It allows methods to be defined directly
// on the slice. This allows the methods to range over the slice instead of
// ranging over the slice in multiple places calling the method.
type Things []*Thing

// Free invokes Thing.Free on each of the *Thing elements in the receiver.
// After the call all elements of the receiver will be removed resulting in an
// empty slice.
func (t *Things) Free() {
	for x := range *t {
		(*t)[x].Free()
		(*t)[x] = nil
	}
	*t = (*t)[:0]
}
