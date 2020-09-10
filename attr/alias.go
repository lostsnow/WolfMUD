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
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Alias attribute.
func init() {
	internal.AddMarshaler((*Alias)(nil), "alias", "aliases")
}

// BUG(diddymus): Aliases are expected be single words only, otherwise they
// probably won't work correctly and cause all sorts of weird problems and
// behaviour.

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
// As well as an alias a Thing may have one or more qualifiers. A qualifier can
// be used to specify a Thing more specifically. For example:
//
//  GET LONG SWORD
//  GET SHORT SWORD
//
// Here the qualifiers are 'LONG' and 'SHORT'. Qualifiers are defined by
// prefixing an alias with a plus '+' symbol. For example:
//
//  Aliases: +LONG SWORD
//  Aliases: +SHORT SWORD
//
// A qualifier can be bound to a specific alias by following the qualifier with
// a colon ':' and the alias to bind it to. For example:
//
//  Aliases: +WOODEN +SHORT:SWORD SHORTSWORD
//
// This binds the qualifier 'SHORT' to the alias 'SWORD'. The following would
// then be valid:
//
//  GET WOODEN SWORD
//  GET WOODEN SHORT SWORD
//  GET SHORT WOODEN SWORD
//  GET WOODEN SHORTSWORD
//  GET SWORD
//  GET SHORTSWORD
//
// The following would not be validi as 'SHORT' is only bound to 'SWORD' and
// not 'SHORTSWORD':
//
//  GET SHORT SHORTSWORD
//
type Alias struct {
	Attribute
	aliases    map[string]struct{}
	qualifiers map[string]struct{}
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
var (
	_ has.Alias = &Alias{}
)

// NewAlias returns a new Alias attribute initialised with the specified
// aliases and qualifiers. Qualifiers are specified by prefixing an alias with
// a plus '+' symbol. The specified aliases and qualifiers are automatically
// uppercased when stored. A unique alias using the parent Thing.UID will be
// added automatically.
func NewAlias(aliases ...string) *Alias {
	a := make(map[string]struct{}, len(aliases))
	q := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		// Ignore empty aliases and qualifiers
		if len(alias) == 0 || len(alias) == 1 && alias == "+" {
			continue
		}
		// Store uppercased alias/qualifier. For qualifiers drop leading '+' before
		// storing.
		if alias[0] != '+' {
			a[strings.ToUpper(alias)] = struct{}{}
		} else {
			q[strings.ToUpper(alias[1:])] = struct{}{}
			if s := strings.SplitAfter(alias, ":"); len(s) == 2 {
				a[strings.ToUpper(s[1])] = struct{}{}
			}
		}
	}
	return &Alias{Attribute{}, a, q}
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
	if t != nil {
		if uid := t.UID(); len(uid) != 0 {
			a.aliases[t.UID()] = struct{}{}
		}
	}
	a.Attribute.SetParent(t)
}

// FindAlias searches the attributes of the specified Thing for attributes that
// implement has.Alias returning the first match it finds or a *Alias typed nil
// otherwise.
func FindAlias(t has.Thing) has.Alias {
	return t.FindAttr((*Alias)(nil)).(has.Alias)
}

// Is returns true if passed attribute implements an alias else false.
func (*Alias) Is(a has.Attribute) bool {
	_, ok := a.(has.Alias)
	return ok
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
	for qualifier := range a.qualifiers {
		aliases = append(aliases, "+"+qualifier)
	}
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

// Dump adds attribute information to the passed tree.Node for debugging.
func (a *Alias) Dump(node *tree.Node) *tree.Node {

	// Format list of aliases and qualifiers as '"A", "B", "C"'
	aliases, qualifiers := []byte{}, []byte{}
	if len(a.aliases) > 0 {
		for l := range a.aliases {
			aliases = strconv.AppendQuote(append(aliases, ", "...), l)
		}
		aliases = aliases[2:]
	}
	if len(a.qualifiers) > 0 {
		for l := range a.qualifiers {
			qualifiers = strconv.AppendQuote(append(qualifiers, ", "...), l)
		}
		qualifiers = qualifiers[2:]
	}

	return node.Append(
		"%p %[1]T - aliases: %d [%s], qualifiers: %d [%s]",
		a, len(a.aliases), aliases, len(a.qualifiers), qualifiers,
	)
}

// HasAlias checks the passed string for a matching alias. Returns true if a
// match is found otherwise false.
func (a *Alias) HasAlias(alias string) (found bool) {
	if a != nil {
		_, found = a.aliases[alias]
	}
	return
}

// HasQualifier checks the passed string for a matching qualifier. Returns true
// if a match is found otherwise false.
func (a *Alias) HasQualifier(qualifier string) (found bool) {
	if a != nil {
		_, found = a.qualifiers[qualifier]
	}
	return
}

// HasQualifierForAlias checks the passed string for a matching qualifier for a
// specific alias. Returns true if a match is found otherwise false.
// An alias can be bound to a qualifier using the syntax "+qualifier:alias".
func (a *Alias) HasQualifierForAlias(alias, qualifier string) (found bool) {
	if a != nil {
		_, found = a.qualifiers[qualifier+":"+alias]
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

// Qualifiers returns a []string of all the qualifiers for an Alias attribute.
// If there are no qualifiers an empty slice will be returned.
func (a *Alias) Qualifiers() (qualifiers []string) {
	if a != nil {
		for qualifier := range a.qualifiers {
			qualifiers = append(qualifiers, qualifier)
		}
	}
	return
}

// Copy returns a copy of the Alias receiver.
func (a *Alias) Copy() has.Attribute {
	if a == nil {
		return (*Alias)(nil)
	}
	aliases := a.Aliases()
	for q := range a.qualifiers {
		aliases = append(aliases, "+"+q)
	}
	return NewAlias(aliases...)
}
