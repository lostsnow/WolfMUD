// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package location

import (
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/utils/command"
	"code.wolfmud.org/WolfMUD.git/utils/inventory"
	"code.wolfmud.org/WolfMUD.git/utils/messaging"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"
	"code.wolfmud.org/WolfMUD.git/utils/text"

	"fmt"
	"strings"
)

const (
	CROWD_SIZE = 10 // How many mobiles make a crowd
)

// Basic provides a default location implementation
//
// NOTE: We could save memory at the cost of performance by using an Inventory
// pointer and not allocating it until something is added - via Add. We could
// also set it to nil when the last Thing is removed - via Remove. Performance
// wise would incur a penality creating the Inventory and also create more for
// the GC to handle... worth investigation? Maybe a pool of inventories?
type Basic struct {
	thing.Thing
	inventory.Inventory
	directionalExits
	mutex chan bool
}

// Register zero value instance of Basic with the loader.
func init() {
	recordjar.RegisterUnmarshaler("basic", &Basic{})
}

// Unmarshal takes a recordjar.Record and allocates the data in it to the passed
// Basic type.
func (b *Basic) Unmarshal(d recordjar.Decoder) {
	b.Thing.Unmarshal(d)
	b.mutex = make(chan bool, 1)
	b.mutex <- true
}

func (b *Basic) Init(d recordjar.Decoder, refs map[string]recordjar.Unmarshaler) {
	b.Thing.Init(d, refs)

	// Link directional exits to locations via direction/location reference pairs
	// e.g. Exits: E→L2 S→L3 NW→L4
	for _, pair := range d.PairList("exits") {

		dir, loc := pair[0], pair[1]

		if l, ok := refs[loc].(Interface); ok {
			if dirIndex, ok := directionShortIndex[dir]; ok {
				b.LinkExit(dirIndex, l)
			}
		}
	}

}

// LinkExit links one location to another in the direction given. This is
// normally only done at setup time when the locations are initilized.
//
// NOTE: The Java version had softlinking - is it still needed?
func (b *Basic) LinkExit(d direction, to Interface) {
	b.directionalExits[d] = to
}

// Add puts a Thing at this location.
func (b *Basic) Add(thing thing.Interface) {
	if t, ok := thing.(Locateable); ok {
		t.Relocate(b)
	}
	b.Inventory.Add(thing)
}

// Remove takes a Thing from this location.
func (b *Basic) Remove(thing thing.Interface) {
	if t, ok := thing.(Locateable); ok {
		t.Relocate(nil)
	}
	b.Inventory.Remove(thing)
}

// Broadcast sends a message to all responders at this location. This
// implements the broadcast.Interface - see that for more details.
func (b *Basic) Broadcast(omit []thing.Interface, format string, any ...interface{}) {
	msg := text.Colorize(fmt.Sprintf("\n"+format, any...))

	for _, item := range b.Inventory.List(omit...) {
		switch messenger := item.(type) {
		case messaging.Responder:
			messenger.Respond(msg)
		case messaging.Broadcaster:
			messenger.Broadcast(omit, format, any...)
		}
	}
}

// Process implements the command.Interface to handle location specific
// commands. First we see if anything at the location can process the command
// and then the location itself. By handling commands in this order anything at
// a location: doors, barriers, guards, etc - can effect movement easily.
func (b *Basic) Process(cmd *command.Command) (handled bool) {

	if handled = b.Inventory.Process(cmd); handled {
		return
	}

	// The following commands can only be processed at the issuer's location. So
	// we need to check if this location is where the issuer is.
	if l, ok := cmd.Issuer.(Locateable); ok {
		if !l.Locate().IsAlso(b) {
			return
		}
	}

	switch cmd.Verb {
	case "LOOK", "L":
		handled = b.look(cmd)
	case "EXITS", "EX":
		handled = b.exits(cmd)
	case "NORTH", "N":
		handled = b.move(cmd, NORTH)
	case "NORTHEAST", "NE":
		handled = b.move(cmd, NORTHEAST)
	case "EAST", "E":
		handled = b.move(cmd, EAST)
	case "SOUTHEAST", "SE":
		handled = b.move(cmd, SOUTHEAST)
	case "SOUTH", "S":
		handled = b.move(cmd, SOUTH)
	case "SOUTHWEST", "SW":
		handled = b.move(cmd, SOUTHWEST)
	case "WEST", "W":
		handled = b.move(cmd, WEST)
	case "NORTHWEST", "NW":
		handled = b.move(cmd, NORTHWEST)
	case "UP", "U":
		handled = b.move(cmd, UP)
	case "DOWN", "D":
		handled = b.move(cmd, DOWN)
	}

	return
}

// BUG(Diddymus): The Java version listed mobiles before other things in look.

// look implements the 'LOOK' command. It describes the location displaying the
// title, description, things and directional exits.
//
// TODO: Implement brief mode.
//
// TODO: Implement looking in a specific direction with a maximum viewing
// distance.
func (b *Basic) look(cmd *command.Command) (handled bool) {

	things := ""

	if b.Crowded() {
		things = "[GREEN]You can see a crowd here.\n"
	} else {
		list := b.List(cmd.Issuer)
		thingsHere := make([]string, 0, len(list))
		for _, o := range list {
			thingsHere = append(thingsHere, "You can see "+o.Name()+" here.")
		}
		if len(thingsHere) > 0 {
			things = "[GREEN]" + strings.Join(thingsHere, "\n") + "\n"
		}
	}

	cmd.Respond("[CYAN]%s\n[WHITE]%s\n%s\n%s", b.Name(), b.Description(), things, b.directionalExits)

	return true
}

// exits implements the 'EXITS' command. It displays the currently available
// directional exits from the location.
func (b *Basic) exits(cmd *command.Command) (handled bool) {
	cmd.Respond("%s", b.directionalExits)
	return true
}

// move implements the directional movement commands. This allows movement from
// location to location by typing a direction such as N or North.
//
// TODO: Modify command so that it can handle buffering of multiple location
// broadcasts.
func (b *Basic) move(cmd *command.Command, d direction) (handled bool) {
	if to := b.directionalExits[d]; to != nil {
		if !cmd.CanLock(to) {
			cmd.AddLock(to)
			return true
		}

		b.Remove(cmd.Issuer)

		// If the location is crowded you are not going to notice someone leaving
		if !b.Crowded() {
			b.Broadcast([]thing.Interface{cmd.Issuer}, "[YELLOW]You see %s go %s.", cmd.Issuer.Name(), directionLongNames[d])
		}

		to.Add(cmd.Issuer)

		// If the location is crowded you are not going to notice someone entering
		if !to.Crowded() {
			to.Broadcast([]thing.Interface{cmd.Issuer}, "[YELLOW]You see %s walk in.", cmd.Issuer.Name())
		}

		to.look(cmd)
	} else {
		cmd.Respond("[RED]You can't go %s from here!", directionLongNames[d])
	}
	return true
}

// Lock is a blocking channel lock. It is unlocked by calling Unlock. Unlock
// should only be called when the lock is held via a successful Lock call. The
// reason for the method instead of making the lock in the struct public - you
// cannot access struct properties directly through the Interface.
func (b *Basic) Lock() {
	<-b.mutex
}

// Unlock unlocks a locked Thing. See Lock method for details.
func (b *Basic) Unlock() {
	b.mutex <- true
}

// BUG(Diddymus): The Crowded method currently counts everything in a location.
// Really it should probably only count mobiles.

// Crowded returns wether a locatioin is crowded or not based on CROWD_SIZE and
// the number of things in the location.
func (b *Basic) Crowded() bool {
	return b.Inventory.Length() >= CROWD_SIZE
}
