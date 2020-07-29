// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"strings"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text/tree"
)

// Register marshaler for Vetoes attribute.
func init() {
	internal.AddMarshaler((*Vetoes)(nil), "veto", "vetoes")
}

// TODO: Currently vetoes can only be applied to the Thing they are vetoing
// for. This means, for example, a guard could not veto the get of items at a
// location it is guarding. Also a Veto is static and unconditional.
//
// TODO: Currently a veto cannot be dynamically added or removed.

// Vetoes implement an attribute for lists of Veto preventing commands for a
// Thing that would otherwise be valid. For example you could Veto the drop
// command if a very sticky item is picked up :)
type Vetoes struct {
	Attribute
	vetoes map[string]has.Veto
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
var (
	_ has.Vetoes = &Vetoes{}
)

// NewVetoes returns a new Vetoes attribute initialised with the specified
// Vetos.
func NewVetoes(veto ...has.Veto) *Vetoes {
	vetoes := &Vetoes{Attribute{}, make(map[string]has.Veto)}
	for _, v := range veto {
		vetoes.vetoes[v.Command()] = v
	}
	return vetoes
}

// FindAllVetoes searches the attributes of the specified Thing for attributes
// that implement has.Vetoes returning all that match. If no matches are found
// an empty slice will be returned.
func FindAllVetoes(t has.Thing) (matches []has.Vetoes) {
	vetoes := t.FindAttrs((*Vetoes)(nil))
	matches = make([]has.Vetoes, len(vetoes))
	for a := range vetoes {
		matches[a] = vetoes[a].(has.Vetoes)
	}
	return
}

// Is returns true if passed attribute implements vetoes else false.
func (*Vetoes) Is(a has.Attribute) bool {
	_, ok := a.(has.Vetoes)
	return ok
}

// Found returns false if the receiver is nil otherwise true.
func (v *Vetoes) Found() bool {
	return v != nil
}

// Unmarshal is used to turn the passed data into a new Vetoes attribute.
func (*Vetoes) Unmarshal(data []byte) has.Attribute {
	veto := []has.Veto{}
	for cmd, msg := range decode.KeyedStringList(data) {
		if cmd == "" || msg == "" {
			continue // Ignore incomplete pairs
		}
		veto = append(veto, NewVeto(cmd, msg))
	}
	return NewVetoes(veto...)
}

// Marshal returns a tag and []byte that represents the receiver.
func (v *Vetoes) Marshal() (tag string, data []byte) {

	pairs := map[string]string{}

	for _, veto := range v.vetoes {
		pairs[veto.Command()] = veto.Message()
	}

	if len(v.vetoes) < 2 {
		tag = "veto"
	} else {
		tag = "vetoes"
	}

	return tag, encode.KeyedStringList(pairs, '→')
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (v *Vetoes) Dump(node *tree.Node) *tree.Node {
	node = node.Append("%p %[1]T - vetoes: %d", v, len(v.vetoes))
	branch := node.Branch()
	for _, veto := range v.vetoes {
		veto.Dump(branch)
	}
	return node
}

// Check checks if any of the passed commands, issued by the passed actor, are
// vetoed. The first matching Veto found is returned otherwise nil is returned.
func (v *Vetoes) Check(actor has.Thing, cmd ...string) has.Veto {
	if v == nil {
		return nil
	}

	// For single checks we can take a shortcut
	if len(cmd) == 1 {
		veto, _ := v.vetoes[cmd[0]]
		return veto
	}

	// For multiple checks return the first that is vetoed
	for _, cmd := range cmd {
		if veto, _ := v.vetoes[cmd]; veto != nil {
			return veto
		}
	}
	return nil
}

// Copy returns a copy of the Vetoes receiver.
func (v *Vetoes) Copy() has.Attribute {
	if v == nil {
		return (*Vetoes)(nil)
	}
	nv := make([]has.Veto, 0, len(v.vetoes))
	for _, v := range v.vetoes {
		nv = append(nv, NewVeto(v.Command(), v.Message()))
	}
	return NewVetoes(nv...)
}

// Free makes sure references are nil'ed when the Vetoes attribute is freed.
func (v *Vetoes) Free() {
	if v == nil {
		return
	}
	for cmd := range v.vetoes {
		v.vetoes[cmd] = nil
		delete(v.vetoes, cmd)
	}
	v.Attribute.Free()
}

// Veto implements a veto for a specific command. Veto need to be added to a
// Vetoes list using NewVetoes.
type Veto struct {
	cmd string
	msg string
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
var (
	_ has.Veto = &Veto{}
)

// NewVeto returns a new Veto attribute initialised for the specified command
// with the specified message text. The command is a normal command such as GET
// and DROP and will automatically be uppercased. The message text should
// indicate why the command was vetoed such as "You can't drop the sword. It
// seems to be cursed". Referring to specific items - such as the sword in the
// example - is valid as a Veto is for a specific known Thing.
func NewVeto(cmd string, msg string) *Veto {
	return &Veto{strings.ToUpper(cmd), msg}
}

// Dump adds attribute information to the passed tree.Node for debugging.
func (v *Veto) Dump(node *tree.Node) *tree.Node {
	return node.Append("%p %[1]T - cmd: %q, msg: %q",
		v, v.Command(), v.Message(),
	)
}

// Command returns the command associated with the Veto.
func (v *Veto) Command() string {
	return v.cmd
}

// Message returns the message associated with the Veto.
func (v *Veto) Message() string {
	return v.msg
}
