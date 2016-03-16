// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"

	"strconv"
	"strings"
)

// Register marshaler for Alias attribute.
func init() {
	internal.AddMarshaler((*Alias)(nil), "alias", "aliases")
}

// BUG(diddymus): Aliases are expected be single words only, otherwise they
// probably won't work correctly and cause all sorts of weird problems and
// behaviour. See also TODO about prefixes.

// Alias implements an attribute for referring to a Thing. An alias is a single
// word used to refer to things. Things may have more than one alias. For
// example a sword may have the aliases 'SWORD' and 'SHORTSWORD'. Given these
// aliases a player may use commands such as:
//
//	GET SWORD
//	EXAMINE SHORTSWORD
//	DROP SHORTSWORD
//
// TODO: Need to implement alias prefixes. This would allow us to distinguish
// between two similar items with the same alias. For example if there are two
// coins, one copper and one silver, we could use either "GET COPPER COIN"
// or "GET SILVER COIN". If there is only one coin then "GET COIN" would be
// sufficient. See also BUG about aliases being single words.
type Alias struct {
	Attribute
	aliases map[string]struct{}
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
var (
	_ has.Alias = &Alias{}
)

// NewAlias returns a new Alias attribute initialised with the specified
// aliases. The specified aliases are automatically uppercased when stored.
func NewAlias(aliases ...string) *Alias {
	a := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		a[strings.ToUpper(alias)] = struct{}{}
	}
	return &Alias{Attribute{}, a}
}

// FindAlias searches the attributes of the specified Thing for attributes that
// implement has.Alias returning the first match it finds or a *Alias typed nil
// otherwise.
func FindAlias(t has.Thing) has.Alias {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Alias); ok {
			return a
		}
	}
	return (*Alias)(nil)
}

// Found returns false if the receiver is nil otherwise true.
func (a *Alias) Found() bool {
	return a != nil
}

// Unmarshal is used to turn the passed data into a new Alias attribute.
func (_ *Alias) Unmarshal(data []byte) has.Attribute {
	return NewAlias(recordjar.Decode.KeywordList(data)...)
}

func (a *Alias) Dump() []string {
	buff := []byte{}
	for a := range a.aliases {
		buff = append(buff, ", "...)
		buff = strconv.AppendQuote(buff, a)
	}
	if len(buff) > 0 {
		buff = buff[2:]
	}
	return []string{DumpFmt("%p %[1]T %d aliases: %s", a, len(a.aliases), buff)}
}

// HasAlias checks the passed string for a matching alias. Returns true if a
// match is found otherwise false.
func (a *Alias) HasAlias(alias string) (found bool) {
	if a != nil {
		_, found = a.aliases[alias]
	}
	return
}
