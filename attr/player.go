// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"io"
	"time"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
	"code.wolfmud.org/WolfMUD.git/text"
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
	acct *account
}

// Some interfaces we want to make sure we implement
var (
	_ has.Player = &Player{}
	_ has.Vetoes = &Player{}
)

// NewPlayer returns a new Player attribute initialised with the specified
// Writer which is used to send data back to the associated client.
func NewPlayer(w io.Writer) *Player {
	return &Player{Attribute{}, w, has.StyleBrief, &account{}}
}

func (p *Player) Dump() []string {
	return []string{DumpFmt("%p %[1]T", p)}
}

// FindPlayer searches the attributes of the specified Thing for attributes
// that implement has.Player returning the first match it finds or a *Player
// typed nil otherwise.
func FindPlayer(t has.Thing) has.Player {
	return t.FindAttr((*Player)(nil)).(has.Player)
}

// Is returns true if passed attribute implements a player else false.
func (*Player) Is(a has.Attribute) bool {
	_, ok := a.(has.Player)
	return ok
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
// is mostly useful for dynamic prompts that show player stats such as health.
//
// NOTE: We always take the oppertunity to update stats when building the
// prompt even if they are not included in the output.
func (p *Player) buildPrompt() (prompt []byte) {

	h := FindHealth(p.Parent())

	switch p.PromptStyle {
	case has.StyleNone:
		h.State()
		prompt = append(prompt, text.Prompt...)
	case has.StyleBrief:
		h.State()
		prompt = append(prompt, text.Prompt...)
		prompt = append(prompt, '>')
	case has.StyleShort:
		prompt = append(prompt, "H:"...)
		prompt = append(prompt, h.Prompt(true)...)
		prompt = append(prompt, text.Prompt...)
		prompt = append(prompt, '>')
	case has.StyleLong:
		prompt = append(prompt, "Health: "...)
		prompt = append(prompt, h.Prompt(false)...)
		prompt = append(prompt, ' ')
		prompt = append(prompt, text.Prompt...)
		prompt = append(prompt, '>')
	}

	return prompt
}

// Unmarshal is used to turn the passed data into a new Player attribute. At
// the moment Player attributes are created internally so return an untyped nil
// so we get ignored.
func (*Player) Unmarshal(data []byte) has.Attribute {
	return nil
}

// Marshal returns a tag and []byte that represents the receiver. In this case
// we return empty values as the Player attribute is not persisted.
func (*Player) Marshal() (string, []byte) {
	return "", []byte{}
}

// Write writes the specified byte slice to the associated client.
func (p *Player) Write(b []byte) (n int, err error) {
	if p == nil {
		return
	}

	b = append(b, p.buildPrompt()...)
	n, err = p.Writer.Write(b)
	return
}

// Copy returns a copy of the Player receiver.
//
// NOTE: The copy will use the same io.Writer as the original.
func (p *Player) Copy() has.Attribute {
	if p == nil {
		return (*Player)(nil)
	}
	np := NewPlayer(p.Writer)
	np.SetPromptStyle(p.PromptStyle)
	return np
}

// Free makes sure references are nil'ed when the Player attribute is freed.
func (p *Player) Free() {
	if p != nil {
		p.Writer = nil
		p.Attribute.Free()
	}
}

// Check will always veto a player being junked and trying to use player as a
// container.
func (p *Player) Check(actor has.Thing, cmds ...string) has.Veto {
	for _, cmd := range cmds {
		switch cmd {
		case "JUNK":
			who := text.TitleFirst(FindName(p.Parent()).TheName("Someone"))
			return NewVeto(cmd, who+" does not want to be junked!")
		case "PUTIN":
			who := FindName(p.Parent()).TheName("Someone")
			return NewVeto(cmd, "You can't put anything into "+who+"!")
		}
	}
	return nil
}

// Account returns the account information for a player. This can be used to
// Marshal, Unmarshal or set a player's account information.
func (p *Player) Account() *account {
	return p.acct
}

// account contains information about the player's account. An account only
// contains the hashes for the account id and passwords.
type account struct {
	account  string    // Account hash
	password string    // Password hash
	salt     string    // Printable salt
	created  time.Time // Timestamp account was created
}

// Set new account information for a player account.
func (a *account) Set(ahash, phash, salt string, created time.Time) {
	a.account = ahash
	a.password = phash
	a.salt = salt
	a.created = created
}

// Marshal a player's account information into a recordjar.Record.
func (a *account) Marshal() recordjar.Record {
	return recordjar.Record{
		"account":  encode.String(a.account),
		"password": encode.String(a.password),
		"salt":     encode.String(a.salt),
		"created":  encode.DateTime(a.created),
	}
}

// Unmarshal a recordjar.Record into a player's account information.
func (a *account) Unmarshal(r recordjar.Record) {
	a.account = decode.String(r["ACCOUNT"])
	a.password = decode.String(r["PASSWORD"])
	a.salt = decode.String(r["SALT"])
	a.created = decode.DateTime(r["CREATED"])
}
