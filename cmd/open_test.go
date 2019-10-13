// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd_test

import (
	"testing"
	"time"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/text"
)

// TestOpen_messages checks messages are output in the correct order with the
// correct color as well as being sent to the right players.
func TestOpen_messages(t *testing.T) {

	// Observer Reset+Info shorthand
	const ORI = text.Reset + "\n" + text.Info

	for _, test := range []struct {
		params    string
		actor     string
		observerA string
		observerB string
	}{
		{
			"", // No item or container
			text.Info + "What did you want to open?\n", "", "",
		}, {
			"door", // Single door
			text.Good + "You open the door.\n",
			ORI + "The actor opens a door.\n",
			ORI + "A door opens.\n",
		}, {
			"door", // Single door - duplicate, check world reset
			text.Good + "You open the door.\n",
			ORI + "The actor opens a door.\n",
			ORI + "A door opens.\n",
		}, {
			"token", // Open a non-door held item
			text.Bad + "You see no 'TOKEN' to open.\n", "", "",
		}, {
			"rock", // Open a non-door item at location
			text.Bad + "You cannot open the rock.\n", "", "",
		}, {
			"window", // Open a non-narrative item
			text.Good + "You open the window.\n",
			ORI + "The actor opens a window.\n",
			"",
		}, {
			"trapdoor", // Open something already open
			text.Info + "The trapdoor is already open.\n", "", "",
		},
	} {

		roomA := attr.NewThing(
			attr.NewStart(),
			attr.NewName("Test room A"),
			attr.NewAlias("ROOM_A"),
			attr.NewDescription(
				"This is a room for testing.",
			),
			attr.NewExits(),
			attr.NewInventory(
				attr.NewThing(
					attr.NewName("a door"),
					attr.NewAlias("DOOR"),
					attr.NewDescription("This is a door."),
					attr.NewDoor(attr.East, false, time.Second, 0),
					attr.NewNarrative(),
				),
				attr.NewThing(
					attr.NewName("a window"),
					attr.NewAlias("WINDOW"),
					attr.NewDescription("This is a window."),
					attr.NewDoor(attr.North, false, time.Second, 0),
				),
				attr.NewThing(
					attr.NewName("a trapdoor"),
					attr.NewAlias("TRAPDOOR"),
					attr.NewDescription("This is a wooden trapdoor in the floor."),
					attr.NewDoor(attr.Down, true, time.Second, 0),
					attr.NewNarrative(),
				),
				attr.NewThing(
					attr.NewName("a rock"),
					attr.NewAlias("ROCK"),
					attr.NewDescription("This is a small rock."),
				),
			),
		)

		roomB := attr.NewThing(
			attr.NewName("Test room B"),
			attr.NewAlias("ROOM_B"),
			attr.NewDescription(
				"This is a room for testing.",
			),
			attr.NewExits(),
			attr.NewInventory(),
		)

		world := attr.Things{roomA, roomB}

		// Link Room A east exit to Room B west exit
		attr.FindExits(roomA).AutoLink(attr.East, attr.FindInventory(roomB))

		actor := cmd.NewTestPlayer("an actor", "ACTOR",
			attr.NewThing(
				attr.NewName("a box"),
				attr.NewAlias("CONTAINER", "BOX"),
				attr.NewDescription("This is a box."),
				attr.NewInventory(),
			),
			attr.NewThing(
				attr.NewName("a token"),
				attr.NewAlias("+TEST", "TOKEN"),
				attr.NewDescription("This is a test token."),
			),
		)

		observerA := cmd.NewTestPlayer("observer A", "OBSERVER_A")

		// Create second observer and move to room B - other side of door
		observerB := cmd.NewTestPlayer("observer B", "OBSERVER_B")
		attr.FindInventory(roomA).Move(observerB, attr.FindInventory(roomB))

		c := "open " + test.params
		t.Run(c, func(t *testing.T) {
			cmd.Parse(actor, c)
			if have := actor.Messages(); have != test.actor {
				t.Errorf("Actor for %+q:\nhave: %+q\nwant: %+q", c, have, test.actor)
			}
			if have := observerA.Messages(); have != test.observerA {
				t.Errorf("Observer A for %+q:\nhave: %+q\nwant: %+q", c, have, test.observerA)
			}
			if have := observerB.Messages(); have != test.observerB {
				t.Errorf("Observer B for %+q:\nhave: %+q\nwant: %+q", c, have, test.observerB)
			}
		})

		world.Free()
	}
}
