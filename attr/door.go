// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/event"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar"

	"log"
	"strings"
	"time"
)

// Register marshaler for Door attribute.
func init() {
	internal.AddMarshaler((*Door)(nil), "door")
}

// Door implements an attribute for blocking exits. Doors are the most common
// way of blocking an exit but this attribute may relate to gates, grills,
// bookcases and other such obstacles.
type Door struct {
	Attribute
	direction byte // Exit door blocks (See attr.Exit constants)
	*state
}

// state represents the current state of a Door. It is shared between the
// original Door and the 'other side' Door so that they will open, close and
// reset together.
type state struct {
	reset    time.Duration // Duration until door resets to initial state
	initOpen bool          // Initial state
	open     bool          // Current state
	event.Cancel
}

// Some interfaces we want to make sure we implement
var (
	_ has.Door        = &Door{}
	_ has.Description = &Door{}
	_ has.Vetoes      = &Door{}
)

// NewDoor returns a new Door attribute. The direction is the direction the
// door blocks - specified as per attr.Exit constants. Open specifies whether
// the door is initially open (true) or closed (false). The reset is the
// duration to wait before resetting the door to its initial state - open or
// closed as specified by open.
func NewDoor(direction byte, open bool, reset time.Duration) *Door {
	return &Door{Attribute{}, direction, &state{reset, open, open, nil}}
}

// FindDoor searches the attributes of the specified Thing for attributes that
// implement has.Door returning the first match it finds or a *Door typed nil
// otherwise.
func FindDoor(t has.Thing) has.Door {
	for _, a := range t.Attrs() {
		if a, ok := a.(has.Door); ok {
			return a
		}
	}
	return (*Door)(nil)
}

// Found returns false if the receiver is nil otherwise true.
func (n *Door) Found() bool {
	return n != nil
}

// Unmarshal is used to turn the passed data into a new Door attribute.
func (*Door) Unmarshal(data []byte) has.Attribute {

	door := NewDoor(0, false, time.Duration(0))
	pairs := recordjar.Decode.PairList(data)

	for _, pair := range pairs {
		switch pair[0] {
		case "EXIT":
			e := NewExits()
			door.direction, _ = e.NormalizeDirection(pair[1])
		case "RESET":
			door.reset, _ = time.ParseDuration(strings.ToLower(pair[1]))
		case "OPEN":
			door.initOpen = recordjar.Decode.Boolean([]byte(pair[1]))
			door.open = door.initOpen
		default:
			log.Printf("Door.unmarshal unknown attribute: %q: %q", pair[0], pair[1])
		}
	}
	return door
}

func (d *Door) Dump() (buff []string) {
	e := NewExits()
	buff = append(buff, DumpFmt("%p %[1]T Exit: %q", d, e.ToName(d.direction)))
	for _, line := range d.state.dump() {
		buff = append(buff, DumpFmt("%s", line))
	}
	return
}

func (s *state) dump() (buff []string) {
	buff = append(buff, DumpFmt("%p %[1]T Reset: %q Init: %t Open: %t", s, s.reset, s.initOpen, s.open))
	buff = append(buff, DumpFmt("%p %[1]T", s.Cancel))
	return
}

// Direction returns the direction of the exit being blocked. The returned
// value matches the constants defined in attr.Exits.
func (d *Door) Direction() byte {
	return d.direction
}

func (d *Door) Description() string {
	if d.open {
		return "It is open."
	}
	return "It is closed."
}

// Check will veto passing through a Door dynamically based on the command
// (direction) given and the current state.
func (d *Door) Check(cmd ...string) has.Veto {

	// If door is open we won't veto
	if d.open {
		return nil
	}

	// Do we understand the command as a direction? If not we won't veto
	e := NewExits()
	dir, err := e.NormalizeDirection(cmd[0])
	if err != nil {
		return nil
	}

	// If the command matches the direction we are blocking veto the command
	if dir == d.direction {

		reason := "You cannot go " +
			e.ToName(d.direction) +
			", " +
			FindName(d.Parent()).Name("something") +
			" is blocking your way."

		return NewVeto(cmd[0], reason)
	}

	// Command didn't match the direction we are blocking
	return nil
}

// Opened returns true if the door is currently open else false.
func (d *Door) Opened() bool {
	return d.open
}

// Closed returns true if the door is currently closed else false.
func (d *Door) Closed() bool {
	return !d.open
}

// Open changes a Door state from closed to open. If there is a pending event
// to open the door it will be cancelled. If the door should automatically
// close again an event to "CLOSE <door>" will be queued. If the door is
// already open calling Open does nothing.
func (d *Door) Open() {
	if d.open {
		return
	}

	if d.Cancel != nil {
		close(d.Cancel)
		d.Cancel = nil
	}

	d.open = true

	a := FindAlias(d.Parent())
	if !a.Found() {
		return
	}

	aliases := a.Aliases()
	if len(aliases) == 0 {
		return
	}

	if d.reset != 0 && d.open != d.initOpen {
		d.Cancel = event.Queue(d.Parent(), "CLOSE "+aliases[0], d.reset)
	}
}

// Close changes a Door state from open to closed. If there is a pending event
// to close the door it will be cancelled. If the door should automatically
// open again an event to "OPEN <door>" will be queued. If the door is already
// closed calling Close does nothing.
func (d *Door) Close() {
	if !d.open {
		return
	}

	if d.Cancel != nil {
		close(d.Cancel)
		d.Cancel = nil
	}

	d.open = false

	a := FindAlias(d.Parent())
	if !a.Found() {
		return
	}

	aliases := a.Aliases()
	if len(aliases) == 0 {
		return
	}

	if d.reset != 0 && d.open != d.initOpen {
		d.Cancel = event.Queue(d.Parent(), "OPEN "+aliases[0], d.reset)
	}
}

// Copy returns a copy of the Door receiver. Copy will only copy a specific
// Door not an original and 'other side' pair - they have to be copied
// separately if required.
func (d *Door) Copy() has.Attribute {
	if d == nil {
		return (*Door)(nil)
	}
	return NewDoor(d.direction, d.initOpen, d.reset)
}
