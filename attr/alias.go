// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"strconv"
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
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
// Every Alias attribute that has a parent Thing set will have a unique ID
// equal to the result of calling Alias.Parent().UID(). Therefore a specific,
// unique Thing can be found by unique ID using, for example,
// Alias.HasAlias(Thing.UID()) or Inventory.Search(Thing.UID()).
//
// NOTE: It is important to switch to the unique alias whenever possible,
// especially when scripting, so that the correct Thing is used for commands.
// This avoids picking the wrong Thing when a given alias identifies multiple
// Things. For example if we have a respawnable runestone and we get and drop
// the runestone it will be registered for cleanup. However if we just use the
// alias 'RUNESTONE' either the dropped or respawned runestone could be cleaned
// up. If the respawned runestone is cleaned up we could end up in a loop
// respawning and cleaning up the wrong runestone. Using the unique alias of
// the dropped runestone avoids this situation.
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
// A unique alias using the parent Thing.UID will be added automatically.
func NewAlias(aliases ...string) *Alias {
	a := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		a[strings.ToUpper(alias)] = struct{}{}
	}
	return &Alias{Attribute{}, a}
}

// SetParent overrides the default Attribute.SetParent in order to set a
// unique alias based on the parent Thing unique ID. The alias will be equal
// to the value returned by calling Alias.Parent().UID(). When the parent for
// the attribute changes the old unique identifier is removed (if there is
// one) and the new unique alias added before setting the new parent.
func (a *Alias) SetParent(t has.Thing) {
	for alias, _ := range a.aliases {
		if strings.HasPrefix(alias, internal.UIDPrefix) {
			delete(a.aliases, alias)
		}
	}
	if uid := t.UID(); len(uid) != 0 {
		a.aliases[t.UID()] = struct{}{}
	}
	a.Attribute.SetParent(t)
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
func (*Alias) Unmarshal(data []byte) has.Attribute {
	return NewAlias(decode.KeywordList(data)...)
}

// Marshal returns a tag and []byte that represents the receiver.
func (a *Alias) Marshal() (tag string, data []byte) {

	// Make a list of aliases but exclude the unique alias. If we don't then the
	// unique aliases will keep being added to the list.
	uid := a.Parent().UID()
	aliases := []string{}
	for alias := range a.aliases {
		if alias == uid {
			continue
		}
		aliases = append(aliases, alias)
	}

	if len(aliases) < 2 {
		tag = "alias"
	} else {
		tag = "aliases"
	}

	data = encode.KeywordList(aliases)
	return
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

// Aliases returns a []string of all the aliases for an Alias attribute. If
// there are no aliases an empty slice will be returned.
func (a *Alias) Aliases() (aliases []string) {
	if a != nil {
		for alias := range a.aliases {
			aliases = append(aliases, alias)
		}
	}
	return
}

// Copy returns a copy of the Alias receiver.
func (a *Alias) Copy() has.Attribute {
	if a == nil {
		return (*Alias)(nil)
	}
	return NewAlias(a.Aliases()...)
}
