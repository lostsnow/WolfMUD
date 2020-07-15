// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Gender attribute.
func init() {
	internal.AddMarshaler((*Gender)(nil), "gender")
}

// Constants for gender indexes.
const (
	It byte = iota
	Male
	Female
)

// genderName is a lookup table for gender indexes to a gender string.
var genderName = map[byte]string{
	It:     "It",
	Male:   "Male",
	Female: "Female",
}

// genderIndex is a lookup table for gender strings to a gender index.
var genderIndex = map[string]byte{
	"It":     It,
	"MALE":   Male,
	"Male":   Male,
	"male":   Male,
	"M":      Male,
	"m":      Male,
	"FEMALE": Female,
	"Female": Female,
	"female": Female,
	"F":      Female,
	"f":      Female,
}

// Gender implements an attribute for specifying the gender of a Thing. If a
// Thing does not have a specific gender attribute the gender will be the
// non-specific 'It'.
type Gender struct {
	Attribute
	gender byte
}

// Some interfaces we want to make sure we implement
var (
	_ has.Gender = &Gender{}
)

// NewGender returns a gender attribute initialised to the specified gender.
// The gender can be specified using an upper, lower or title cased string. An
// upper or lower case 'M' or 'F' is also understood. If the gender specified
// is not valid the gender will default to a non-specific 'It'.
func NewGender(gender string) *Gender {
	if g, ok := genderIndex[gender]; ok {
		return &Gender{Attribute{}, g}
	}
	return &Gender{Attribute{}, It}
}

// FindGender searches the attributes of the specified Thing for attributes
// that implement has.Gender returning the first match it finds or a *Gender
// typed nil otherwise.
func FindGender(t has.Thing) has.Gender {
	return t.FindAttr((*Gender)(nil)).(has.Gender)
}

// Is returns true if passed attribute implements gender else false.
func (*Gender) Is(a has.Attribute) bool {
	_, ok := a.(has.Gender)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (g *Gender) Found() bool {
	return g != nil
}

// Unmarshal is used to turn the passed data into a new Gender attribute.
func (*Gender) Unmarshal(data []byte) has.Attribute {
	return NewGender(decode.String(data))
}

// Marshal returns a tag and []byte that represents the receiver.
func (g *Gender) Marshal() (tag string, data []byte) {
	return "gender", encode.Keyword(g.Gender())
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (g *Gender) Dump(node *tree.Node) *tree.Node {
	return node.Append("%p %[1]T - %q", g, genderName[g.gender])
}

// Gender returns the gender stored in the attribute as a title-cased string.
// If the receiver is nil a non-specific "It" will be returned.
func (g *Gender) Gender() string {
	if g == nil {
		return genderName[It]
	}
	return genderName[g.gender]
}

// Copy returns a copy of the Gender receiver.
func (g *Gender) Copy() has.Attribute {
	if g == nil {
		return (*Gender)(nil)
	}
	return NewGender(genderName[g.gender])
}
