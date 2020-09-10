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

// TestClose_messages checks messages are output in the correct order with the
// correct color as well as being sent to the right players.
func TestClose_messages(t *testing.T) {

	const OI = "\n" + text.Info  // Observer Info shorthand
	const P = "\n" + text.Prompt // Prompt (StyleNone) shorthand

	for _, test := range []struct {
		params    string
		actor     string
		observerA string
		observerB string
	}{
		{
			"", // No item or container
			text.Info + "What did you want to close?" + P, "", "",
		}, {
			"door", // Single door
			text.Good + "You close the door." + P,
			OI + "The actor closes a door." + P,
			OI + "A door closes." + P,
		}, {
			"door", // Single door - duplicate, check world reset
			text.Good + "You close the door." + P,
			OI + "The actor closes a door." + P,
			OI + "A door closes." + P,
		}, {
			"token", // Close a non-door held item
			text.Bad + "You see no 'TOKEN' here to close." + P, "", "",
		}, {
			"rock", // Close a non-door item at location
			text.Bad + "You cannot close the rock." + P, "", "",
		}, {
			"window", // Close a non-narrative item
			text.Good + "You close the window." + P,
			OI + "The actor closes a window." + P,
			"",
		}, {
			"trapdoor", // Close something already close
			text.Info + "The trapdoor is already closed." + P, "", "",
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
					attr.NewDoor(attr.East, true, time.Second, 0),
					attr.NewNarrative(),
				),
				attr.NewThing(
					attr.NewName("a window"),
					attr.NewAlias("WINDOW"),
					attr.NewDescription("This is a window."),
					attr.NewDoor(attr.North, true, time.Second, 0),
				),
				attr.NewThing(
					attr.NewName("a trapdoor"),
					attr.NewAlias("TRAPDOOR"),
					attr.NewDescription("This is a wooden trapdoor in the floor."),
					attr.NewDoor(attr.Down, false, time.Second, 0),
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

		c := "close " + test.params
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

// TestClose_door checks to make sure closing a door actually closes it.
func TestClose_door(t *testing.T) {

	door := attr.NewDoor(attr.East, false, time.Second, 0)

	locA := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		attr.NewInventory(
			attr.NewThing(
				attr.NewName("a door"),
				attr.NewAlias("DOOR"),
				attr.NewDescription("This is a wooden door."),
				attr.NewNarrative(),
				door,
			),
		),
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR")

	// Try to close the door
	cmd.Parse(actor, "close door")

	// Door should now be closed
	if door.Opened() {
		t.Errorf("%s, %s: was not closed.",
			door.Parent(), attr.FindName(door.Parent()).Name("?"),
		)
	}

	locA.Free()
}
