// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"log"
	"time"

	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/event"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/recordjar/decode"
	"code.wolfmud.org/WolfMUD.git/recordjar/encode"
)

// Register marshaler for Door attribute.
func init() {
	internal.AddMarshaler((*Door)(nil), "door")
}

// Door implements an attribute for blocking exits. Doors are the most common
// way of blocking an exit but this attribute may relate to gates, grills,
// bookcases and other such obstacles.
//
// A complete working door consists of two Thing each with a Door attribute.
// One is the original door and the other is the 'other side'. The original
// Door is added to the location with the exit to be blocked, the 'other side'
// is added to the location the exit to be blocked leads to. Taking the tavern
// entrance in data/zones/zinara.wrj as an example:
//
//             _________________________________________
//            |L3                  |L5                  |
//            |    Tavern          #   Between Tavern   |
//                 Entrance        #   & Bakery
//                                 #
//                 (Door)          #   ('Other Side')
//            |                    #                    |   # = a door
//            |__                __|__                __|
//
//
// Here we have locations L3 (Tavern Entrance) and L5 (Between Tavern *
// Bakery). Between them is the Tavern door. It is defined as:
//
//  %%
//        Ref: L3N1
//  Narrative:
//       Name: the tavern door
//    Aliases: DOOR
//       Door: EXIT→E RESET→1m JITTER→1m
//
//  This is a sturdy wooden door with a simple latch.
//  %%
//
// This adds a Thing representing the door to L3 and blocks the exit going east
// (EXIT→E) to L5. During zone loading and Unmarshaling OtherSide is called on
// the original door Thing. This creates another Thing used for the 'Other
// Side'. It is added to the location found by taking the exit the original
// door is blocking, in this case we are blocking the east exit which leads to
// L5. The 'Other Side' is added to L5 and is setup to block the returning
// exit, in this case west - back to L3. Now in L3 if we do 'EXAMINE DOOR' we
// are examining the original, in L5 we are examining the 'Other Side' which
// appears to be the same door. Because the original and 'Other Side' share
// state be can also issue 'OPEN DOOR' or 'CLOSE DOOR' in either L3 or L5.
//
// When the door is not in it's initial state it will reset after a delay of
// between 1 and 2 minutes. That is, sometime between delay and delay+jitter.
//
// If delay and jitter are both zero the door will not reset automatically.
//
// NOTE: For now a Door attribute should only be added to a Thing with a
// Narrative attribute that is placed at a location. Adding a Door attribute to
// a location directly or to a moveable object will result in odd - possibly
// interesting -  behaviour.
type Door struct {
	Attribute
	direction byte // Exit door blocks (See attr.Exit constants)
	*state
}

// state represents the current state of a Door. It is shared between the
// original Door and the 'other side' Door so that they will open, close and
// reset together.
//
// The otherSide flag is to prevent duplicate door creation. For example assume
// we have locations A and B with a door between them. We initialise the
// locations in the order A then B. We find a door in A and create the 'other
// side' in B. We would now find a door in B and create the 'other side' in A.
type state struct {
	reset     time.Duration // Duration until door resets to initial state
	jitter    time.Duration // Modify reset by up to jitter amount
	initOpen  bool          // Initial state
	open      bool          // Current state
	otherSide bool          // Does door have 'other side' yet?
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
// closed as specified by open. The jitter is a random amount of time to add to
// the reset delay. Adding jitter means the Door will reset with an actual
// delay of between delay and delay+jitter.
//
// This actually only creates one side of a door. To create the 'other side' of
// the door Door.OtherSide should be called.
func NewDoor(direction byte, open bool, reset time.Duration, jitter time.Duration) *Door {
	return &Door{Attribute{}, direction, &state{reset, jitter, open, open, false, nil}}
}

// OtherSide creates the 'other side' of a Door and places it in the World. The
// 'other side' will be placed in the Inventory found by following the exit
// that is being blocked by the original Door. Creating the 'other side' will
// fail if:
//
//  - The original Door attribute has not been added to a Thing
//  - The parent Thing of the original Door is not in an Inventory (e.g. location)
//  - The parent Thing of the Inventory the parent Thing of the Door is in has
//    no Exits
//  - There is no exit in the direction the door is supposed to be blocking
//
// For more details see the attr.Door type.
func (d *Door) OtherSide() {

	// Does door have 'other side' already?
	if d.otherSide {
		return
	}

	// Find parent Thing of original Door
	p := d.Parent()
	if p == nil {
		log.Printf("Door attribute has no parent")
		return
	}

	// Try and get originals proper name
	n := FindName(p).Name("'door'")

	// Create 'other side' of the door as a duplicate Thing
	t := p.Copy()

	// Find the door on the 'other side'
	o := FindDoor(t).(*Door)

	// Share its state with the original door. It is important that the two Door
	// share state so that they open, close and reset together.
	o.state = d.state

	// Mark door as having an 'other side'
	o.otherSide = true

	// Point the 'other side' of the door in the opposing direction
	o.direction = Return(d.direction)

	// Find out where the original door is
	w := FindLocate(d.Parent()).Where()
	if w == nil {
		log.Printf("Parent of door %q is nowhere, cannot create other side", n)
		return
	}

	// Find exits for where the door is
	e := FindExits(w.Parent())
	if !e.Found() {
		log.Printf("Parent of door %q has no exits, cannot create other side", n)
		return
	}

	// Find opposing location's inventory
	i := e.LeadsTo(d.direction)
	if i == nil {
		log.Printf("There is no exit %q for Door %q to block, cannot add other side", e.ToName(d.direction), n)
		return
	}

	// Add 'other side' to opposing location's inventory and enable it
	i.Add(t)
	i.Enable(t)

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

	door := NewDoor(0, false, time.Duration(0), time.Duration(0))

	for field, data := range decode.PairList(data) {
		bdata := []byte(data)
		switch field {
		case "EXIT":
			e := NewExits()
			door.direction, _ = e.NormalizeDirection(data)
		case "RESET":
			door.reset = decode.Duration(bdata)
		case "JITTER":
			door.jitter = decode.Duration(bdata)
		case "OPEN":
			door.initOpen = decode.Boolean(bdata)
			door.open = door.initOpen
		default:
			log.Printf("Door.unmarshal unknown attribute: %q: %q", field, data)
		}
	}
	return door
}

// Marshal returns a tag and []byte that represents the receiver.
func (d *Door) Marshal() (tag string, data []byte) {
	tag = "door"
	data = encode.PairList(
		map[string]string{
			"exit":   string(NewExits().ToName(d.direction)),
			"reset":  string(encode.Duration(d.reset)),
			"jitter": string(encode.Duration(d.jitter)),
			"open":   string(encode.Boolean(d.initOpen)),
		},
		'→',
	)
	return
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
	buff = append(buff, DumpFmt("%p %[1]T Reset: %q Jitter: %q Init: %t Open: %t", s, s.reset, s.jitter, s.initOpen, s.open))
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
// (direction) given and the current state of the Door - open or closed.
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

	if d.reset+d.jitter != 0 && d.open != d.initOpen {
		t := d.Parent()
		d.Cancel = event.Queue(t, "CLOSE "+t.UID(), d.reset, d.jitter)
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

	if d.reset+d.jitter != 0 && d.open != d.initOpen {
		t := d.Parent()
		d.Cancel = event.Queue(t, "OPEN "+t.UID(), d.reset, d.jitter)
	}
}

// Copy returns a copy of the Door receiver. Copy will only copy a specific
// Door not an original and 'other side' pair - they have to be copied
// separately if required.
func (d *Door) Copy() has.Attribute {
	if d == nil {
		return (*Door)(nil)
	}
	return NewDoor(d.direction, d.initOpen, d.reset, d.jitter)
}

// Free makes sure references are nil'ed and channels closed when the Door
// attribute is freed.
func (d *Door) Free() {
	if d == nil {
		return
	}
	if d.Cancel != nil {
		close(d.Cancel)
		d.Cancel = nil
	}
	d.state = nil
	d.Attribute.Free()
}
