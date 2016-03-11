// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/has"

	"net"
)

// Register marshaler for Player attribute.
func init() {
	internal.AddMarshaler((*Player)(nil), "player")
}

// Player implements an attribute for associating a thing with a client network
// connection.
type Player struct {
	Attribute
	conn net.Conn
}

// Some interfaces we want to make sure we implement
var (
	_ has.Player = &Player{}
)

// NewPlayer returns a new Player attribute initialised with the specified
// network connection.
func NewPlayer(c net.Conn) *Player {
	return &Player{Attribute{}, c}
}

func (p *Player) Dump() []string {
	return []string{DumpFmt("%p %[1]T %q", p, p.conn.RemoteAddr())}
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

// Unmarshal is used to turn the passed data into a new Player attribute. At
// the moment Player attributes are created internally so return an untyped nil
// so we get ignored.
func (_ *Player) Unmarshal(data []byte) has.Attribute {
	return nil
}

// Write writes the specified byte slice to the network connection associated
// with the Player receiver.
func (p *Player) Write(b []byte) {
	if p != nil {
		p.conn.Write(b)
	}
}
