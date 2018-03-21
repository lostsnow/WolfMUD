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
	t := &Thing{uid: <-internal.NextUID}

	t.Add(a...)

	c := <-ThingCount
	c++
	ThingCount <- c

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
	for i := range t.attrs {
		t.attrs[i].Free()
		t.attrs[i] = nil
	}
	t.attrs = nil
	t.rwmutex.Unlock()

	c := <-ThingCount
	c--
	ThingCount <- c
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

// Attrs returns all of the Attributes a Thing has as a slice of has.Attribute.
// This is commonly used to range over all of the Attributes of a Thing instead
// of using a finder for a specific type of Attribute.
func (t *Thing) Attrs() []has.Attribute {
	t.rwmutex.RLock()
	a := make([]has.Attribute, len(t.attrs))
	for i := range t.attrs {
		a[i] = t.attrs[i]
	}
	t.rwmutex.RUnlock()
	return a
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

	rec := recordjar.Record{}
	rec["ref"] = []byte(t.UID())
	for _, a := range t.Attrs() {
		tag, data = a.Marshal()
		if tag == "" {
			continue
		}
		rec[tag] = data
	}
	return rec
}

func (t *Thing) Dump() (buff []string) {
	t.rwmutex.RLock()
	buff = append(buff, DumpFmt("%s, %d attributes:", t, len(t.attrs)))
	for _, a := range t.attrs {
		for _, a := range a.Dump() {
			buff = append(buff, DumpFmt("%s", a))
		}
	}
	t.rwmutex.RUnlock()
	return buff
}

func DumpFmt(format string, args ...interface{}) string {
	return "  " + fmt.Sprintf(format, args...)
}

// Copy returns a copy of the Thing receiver. The copy will be made recursively
// copying all associated Attribute and Thing.
func (t *Thing) Copy() has.Thing {
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

	// Set the origin for items in our Inventory including disabled items
	for _, t := range append(i.Contents(), i.Narratives()...) {
		if l := FindLocate(t); l.Found() {
			l.SetOrigin(i)
		}
		t.SetOrigins()
	}
	for _, t := range i.Disabled() {
		if l := FindLocate(t); l.Found() {
			l.SetOrigin(i)
		}
		t.SetOrigins()
	}
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

// String causes a Thing to implement the Stringer interface so that a Thing
// can print information about itself. The format of the string is:
//
//  <address> <type> UID: <unique ID>
//
//  0xc420108630 *attr.Thing UID: #UID-6M
//
func (t *Thing) String() string {
	return fmt.Sprintf("%p %[1]T %s", t, t.UID())
}
