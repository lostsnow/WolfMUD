// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"

	"io"
)

// Register marshaler for Player attribute.
func init() {
	internal.AddMarshaler((*Player)(nil), "player")
}

// Player implements an attribute for associating a Thing with a Writer used to
// return data to the associated client.
type Player struct {
	Attribute
	io.Writer
	has.PromptStyle
}

// Some interfaces we want to make sure we implement
var (
	_ has.Player = &Player{}
)

// NewPlayer returns a new Player attribute initialised with the specified
// Writer which is used to send data back to the associated client.
func NewPlayer(w io.Writer) *Player {
	return &Player{Attribute{}, w, has.StyleBrief}
}

func (p *Player) Dump() []string {
	return []string{DumpFmt("%p %[1]T", p)}
}

// FindPlayer searches the attributes of the specified Thing for attributes
// that implement has.Player returning the first match it finds or a *Player
// typed nil otherwise.
func FindPlayer(t has.Thing) has.Player {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Player); ok {
			return a
		}
	}
	return (*Player)(nil)
}

// Found returns false if the receiver is nil otherwise true.
func (p *Player) Found() bool {
	return p != nil
}

// SetPromptStyle is used to set the current prompt style and returns the
// previous prompt style. This is so the previous prompt style can be restored
// if required later on.
func (p *Player) SetPromptStyle(new has.PromptStyle) (old has.PromptStyle) {
	old, p.PromptStyle = p.PromptStyle, new
	return
}

// buildPrompt creates a prompt appropriate for the current PromptStyle. This
// is mostly useful for dynamic prompts that show player statistics.
func (p *Player) buildPrompt() []byte {
	switch p.PromptStyle {
	case has.StyleBrief:
		return []byte(">")
	default:
		return []byte{}
	}
}

// Unmarshal is used to turn the passed data into a new Player attribute. At
// the moment Player attributes are created internally so return an untyped nil
// so we get ignored.
func (*Player) Unmarshal(data []byte) has.Attribute {
	return nil
}

// Write writes the specified byte slice to the associated client.
func (p *Player) Write(b []byte) (n int, err error) {
	if p == nil {
		return
	}

	if len(b) > 0 {
		b = append(b, '\n')
	}
	b = append(b, p.buildPrompt()...)
	if p != nil {
		n, err = p.Writer.Write(b)
	}
	return
}
