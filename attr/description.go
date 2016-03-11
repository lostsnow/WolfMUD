// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"
)

// Register marshaler for Description attribute.
func init() {
	internal.AddMarshaler((*Description)(nil), "description")
}

// Description implements an attribute for describing Things. Things can have
// multiple descriptions or other attributes that implement the has.Description
// interface to add additional information to descriptions.
type Description struct {
	Attribute
	description string
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
var (
	_ has.Description = &Description{}
)

// NewDescription returns a new Description attribute initialised with the
// specified description.
func NewDescription(description string) *Description {
	return &Description{Attribute{}, description}
}

// FindAllDescription searches the attributes of the specified Thing for
// attributes that implement has.Description returning all that match. If no
// matches are found an empty slice will be returned.
func FindAllDescription(t has.Thing) (matches []has.Description) {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Description); ok {
			matches = append(matches, a)
		}
	}
	return
}

// Found returns false if the receiver is nil otherwise true.
func (d *Description) Found() bool {
	return d != nil
}

// Unmarshal is used to turn the passed data into a new Description attribute.
func (_ *Description) Unmarshal(data []byte) has.Attribute {
	return NewDescription(recordjar.Decode.String(data))
}

func (d *Description) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", d, d.description)}
}

// Description returns the descriptive string of the attribute.
func (d *Description) Description() string {
	return d.description
}
