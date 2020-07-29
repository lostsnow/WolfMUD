// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"log"
	"strconv"
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Barrier attribute.
func init() {
	internal.AddMarshaler((*Barrier)(nil), "barrier")
}

// Barrier implements an attribute for conditionally vetoing movement in a
// given direction. The barrier may be invisible or invisible and interactive
// or non-interactive depending on the Thing the attribute is attached to.
// A Barrier has two lists of aliases: those allowed to pass through the
// barrier and those not allowed to pass through the barrier. The allowed
// aliases are checked before those in the denied list. For example:
//
//  b := NewBarrier(East, []string{"GUARD"}, []{"NPC"})
//
// This would define a barrier preventing all mobiles with an alias of "NPC"
// from going east, unless they also has an alias of "GUARD". In this respect
// "NPC" is acting as a group alias, and can be applied to all mobiles that are
// considered non-player characters. If a mobile (or player) does not match an
// alias in the denied list they are allowed to pass through.
type Barrier struct {
	Attribute
	direction byte // Exit the barrier blocks (See attr.Exit constants)

	// NOTE: If allow and deny were []string it would make the code simpler and
	// save swapping between slices and maps - but the common case of lookups in
	// Check would be slower.
	allow map[string]struct{}
	deny  map[string]struct{}
}

// Some interfaces we want to make sure we implement
var (
	_ has.Barrier = &Barrier{}
	_ has.Vetoes  = &Barrier{}
)

// NewBarrier returns a new Barrier attribute. The direction is the
// conditionally blocked exit's direction. The allow and deny slices are lists
// of aliases that decide if movement in the given direction is vetoed or not.
func NewBarrier(direction byte, allow []string, deny []string) *Barrier {
	b := &Barrier{
		Attribute{},
		direction,
		make(map[string]struct{}, len(allow)),
		make(map[string]struct{}, len(deny)),
	}
	for _, v := range allow {
		b.allow[strings.ToUpper(v)] = struct{}{}
	}
	for _, v := range deny {
		b.deny[strings.ToUpper(v)] = struct{}{}
	}
	return b
}

// FindBarrier searches the attributes of the specified Thing for attributes
// that implement has.Barrier returning the first match it finds or a *Barrier
// typed nil otherwise.
func FindBarrier(t has.Thing) has.Barrier {
	return t.FindAttr((*Barrier)(nil)).(has.Barrier)
}

// Is returns true if passed attribute implements a barrier else false.
func (*Barrier) Is(a has.Attribute) bool {
	_, ok := a.(has.Barrier)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (b *Barrier) Found() bool {
	return b != nil
}

// Unmarshal is used to turn the passed data into a new Barrier attribute.
func (*Barrier) Unmarshal(data []byte) has.Attribute {

	var direction byte
	var allow, deny []string

	for field, data := range decode.PairList(data) {
		bdata := []byte(data)
		switch field {
		case "EXIT":
			e := NewExits()
			direction, _ = e.NormalizeDirection(data)
		case "ALLOW":
			allow = strings.Split(decode.Keyword(bdata), ",")
		case "DENY":
			deny = strings.Split(decode.Keyword(bdata), ",")
		default:
			log.Printf("Barrier.unmarshal unknown attribute: %q: %q", field, data)
		}
	}
	return NewBarrier(direction, allow, deny)
}

// Marshal returns a tag and []byte that represents the receiver.
func (b *Barrier) Marshal() (tag string, data []byte) {
	tag = "barrier"
	data = encode.PairList(
		map[string]string{
			"exit":  string(NewExits().ToName(b.direction)),
			"allow": string(encode.Keyword(strings.Join(b.Allowed(), ","))),
			"deny":  string(encode.Keyword(strings.Join(b.Denied(), ","))),
		},
		'→',
	)
	return
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (b *Barrier) Dump(node *tree.Node) *tree.Node {
	e := NewExits()

	// Format list of aliases and qualifiers as '"A", "B", "C"'
	allow, deny := []byte{}, []byte{}
	if len(b.allow) > 0 {
		for l := range b.allow {
			allow = strconv.AppendQuote(append(allow, ", "...), l)
		}
		allow = allow[2:]
	}
	if len(b.deny) > 0 {
		for l := range b.deny {
			deny = strconv.AppendQuote(append(deny, ", "...), l)
		}
		deny = deny[2:]
	}

	return node.Append("%p %[1]T - exit: %q, allow: %d [%s], deny: %d [%s]",
		b, e.ToName(b.direction), len(b.allow), allow, len(b.deny), deny,
	)
}

// Allowed returns a string slice of aliases that can unconditionally pass
// through the barrier.
func (b *Barrier) Allowed() (aliases []string) {
	if b != nil {
		for v := range b.allow {
			aliases = append(aliases, v)
		}
	}
	return
}

// Denied returns a string slice of aliases that can cannot pass through the
// barrier, unless overridden unconditionally by allowed aliases.
func (b *Barrier) Denied() (aliases []string) {
	if b != nil {
		for v := range b.deny {
			aliases = append(aliases, v)
		}
	}
	return
}

// Check will veto passing through a Barrier dynamically based on the command
// (direction) given and the aliases of the actor wishing to pass through the
// Barrier. If an alias of the actor matches a denied alias we veto movement,
// unless overridden by a matching allowed alias. Otherwise we don't veto.
func (b *Barrier) Check(actor has.Thing, cmd ...string) has.Veto {

	// Do we understand the command as a direction? If not we won't veto
	e := NewExits()
	dir, err := e.NormalizeDirection(cmd[0])
	if err != nil {
		return nil
	}

	// Bail if command does not match the direction we are blocking
	if dir != b.direction {
		return nil
	}

	alias := FindAlias(actor)

	// Allow by default if no aliases - deny can never match
	if !alias.Found() {
		return nil
	}

	aliases := alias.Aliases()
	denied := false

	// Check actor aliases against allow/deny lists
	for _, alias := range aliases {
		// Bail early if explicity allowed
		if _, ok := b.allow[alias]; ok {
			return nil
		}
		// Flag a deny match - may be overridden by an allow
		if _, ok := b.deny[alias]; ok {
			denied = true
		}
	}

	// If we get here and a denied was flagged then it wasn't overridden by an
	// allow
	if denied {
		dirName := e.ToName(b.direction)
		what := b.Parent()
		name := "something"

		// If Thing is not a location use the name of the Thing
		if !FindExits(what).Found() {
			name = FindName(what).Name(name)
		}

		return NewVeto(
			cmd[0],
			"You cannot go "+dirName+", "+name+" is blocking your way.",
		)
	}

	// If not explicitly denied or allowed then allow
	return nil
}

// Copy returns a copy of the Barrier receiver.
func (b *Barrier) Copy() has.Attribute {
	if b == nil {
		return (*Barrier)(nil)
	}
	return NewBarrier(b.direction, b.Allowed(), b.Denied())
}
