// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package location

import (
	"fmt"
	"strings"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
	"wolfmud.org/utils/inventory"
	"wolfmud.org/utils/messaging"
	"wolfmud.org/utils/text"
)

// Basic provides a default location implementation
type Basic struct {
	thing.Thing
	inventory.Inventory
	directionalExits
}

// NewBasic creates a new Basic location and returns a reference to it.
//
// NOTE: We could save memory at the cost of performance by not allocating the
// Inventory until something is added - via Add. We could also set it to nil
// when the last Thing is removed - via Remove. Performance wise we would incur
// a penality creating the Inventory and also create a lot more for the GC to
// handle?
func NewBasic(name string, aliases []string, description string) *Basic {
	return &Basic{
		Thing:     *thing.New(name, aliases, description),
		Inventory: *inventory.New(),
	}
}

// LinkExit links one location to another in the direction given. This is
// normally only done at setup time when the world is initially loaded.
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

	if handled = b.Inventory.Delegate(cmd); handled {
		return
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

	list := b.Inventory.List(cmd.Issuer)
	thingsHere := make([]string, 0, len(list))
	for _, o := range list {
		thingsHere = append(thingsHere, "You can see "+o.Name()+" here.")
	}

	things := ""
	if len(thingsHere) > 0 {
		things = strings.Join(thingsHere, "\n") + "\n"
	}

	cmd.Respond("[CYAN]%s[WHITE]\n%s\n[GREEN]%s\n[CYAN]You can see exits: [YELLOW]%s", b.Name(), b.Description(), things, b.directionalExits)

	return true
}

// exits implements the 'EXITS' command. It displays the currently available
// directional exits from the location.
func (b *Basic) exits(cmd *command.Command) (handled bool) {
	cmd.Respond("[CYAN]You can see exits: [YELLOW]%s", b.directionalExits)
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
		b.Broadcast([]thing.Interface{cmd.Issuer}, "[YELLOW]You see %s go %s.", cmd.Issuer.Name(), directionNames[d])

		to.Add(cmd.Issuer)
		to.Broadcast([]thing.Interface{cmd.Issuer}, "[YELLOW]You see %s walk in.", cmd.Issuer.Name())

		to.look(cmd)
	} else {
		cmd.Respond("You can't go %s from here!", directionNames[d])
	}
	return true
}
