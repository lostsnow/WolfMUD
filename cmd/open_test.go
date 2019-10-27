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
		observerA string // Room A with Actor
		observerB string // Room B
		observerC string // Room C
	}{
		{
			"", // No item or container
			text.Info + "What did you want to open?\n", "", "", "",
		}, {
			"door", // Single door
			text.Good + "You open the red door.\n",
			ORI + "The actor opens a red door.\n",
			ORI + "A red door opens.\n",
			"",
		}, {
			"door", // Single door - duplicate, check world reset
			text.Good + "You open the red door.\n",
			ORI + "The actor opens a red door.\n",
			ORI + "A red door opens.\n",
			"",
		}, {
			"red door", // Door with specific qualifier to east
			text.Good + "You open the red door.\n",
			ORI + "The actor opens a red door.\n",
			ORI + "A red door opens.\n",
			"",
		}, {
			"blue door", // Door with specific qualifier to west
			text.Good + "You open the blue door.\n",
			ORI + "The actor opens a blue door.\n",
			"",
			ORI + "A blue door opens.\n",
		}, {
			"wooden door", // First door matching multiple qualifier
			text.Good + "You open the red door.\n",
			ORI + "The actor opens a red door.\n",
			ORI + "A red door opens.\n",
			"",
		}, {
			"2nd door", // Specific instance of a door
			text.Good + "You open the blue door.\n",
			ORI + "The actor opens a blue door.\n",
			"",
			ORI + "A blue door opens.\n",
		}, {
			"2nd wooden door", // Specific instance with qualifier
			text.Good + "You open the blue door.\n",
			ORI + "The actor opens a blue door.\n",
			"",
			ORI + "A blue door opens.\n",
		}, {
			"green door", // Door with unknown qualifier, fails match
			text.Bad + "You see no 'GREEN DOOR' here to open.\n", "", "", "",
		}, {
			"all door", // More than one door specified
			text.Bad + "You can only open one thing at a time.\n", "", "", "",
		}, {
			"all wooden door", // More than one qualified door specified
			text.Bad + "You can only open one thing at a time.\n", "", "", "",
		}, {
			"red door blue door", // Valid match, but unmatched words, so fail match
			text.Bad + "You see no 'RED DOOR BLUE DOOR' here to open.\n", "", "", "",
		}, {
			"frog turtle", // Invalid items
			text.Bad + "You see no 'FROG TURTLE' here to open.\n", "", "", "",
		}, {
			"frog door turtle", // Invalid and valid items mixed
			text.Bad + "You see no 'FROG DOOR TURTLE' here to open.\n",
			"", "", "",
		}, {
			"token", // Open a non-door held item
			text.Bad + "You see no 'TOKEN' here to open.\n", "", "", "",
		}, {
			"rock", // Open a non-door item at location
			text.Bad + "You cannot open the rock.\n", "", "", "",
		}, {
			"window", // Open a non-narrative item
			text.Good + "You open the window.\n",
			ORI + "The actor opens a window.\n",
			"",
			"",
		}, {
			"trapdoor", // Open something already open
			text.Info + "The trapdoor is already open.\n", "", "", "",
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
					attr.NewName("a red door"),
					attr.NewAlias("+WOODEN", "+RED", "DOOR"),
					attr.NewDescription("This is a red, wooden door."),
					attr.NewDoor(attr.East, false, time.Second, 0),
					attr.NewNarrative(),
				),
				attr.NewThing(
					attr.NewName("a blue door"),
					attr.NewAlias("+WOODEN", "+BLUE", "DOOR"),
					attr.NewDescription("This is a blue, wooden door."),
					attr.NewDoor(attr.West, false, time.Second, 0),
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

		roomC := attr.NewThing(
			attr.NewName("Test room C"),
			attr.NewAlias("ROOM_C"),
			attr.NewDescription(
				"This is a room for testing.",
			),
			attr.NewExits(),
			attr.NewInventory(),
		)

		world := attr.Things{roomA, roomB, roomC}

		// Link Room A east exit to Room B west exit
		attr.FindExits(roomA).AutoLink(attr.East, attr.FindInventory(roomB))

		// Link Room A west exit to Room C east exit
		attr.FindExits(roomA).AutoLink(attr.West, attr.FindInventory(roomC))

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

		// Create second observer and move to room B - other side of 1st door
		observerB := cmd.NewTestPlayer("observer B", "OBSERVER_B")
		attr.FindInventory(roomA).Move(observerB, attr.FindInventory(roomB))

		// Create third observer and move to room C - other side of 2nd door
		observerC := cmd.NewTestPlayer("observer C", "OBSERVER_C")
		attr.FindInventory(roomA).Move(observerC, attr.FindInventory(roomC))

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
			if have := observerC.Messages(); have != test.observerC {
				t.Errorf("Observer C for %+q:\nhave: %+q\nwant: %+q", c, have, test.observerC)
			}
		})

		world.Free()
	}
}

// TestOpen_door checks to make sure opening a door actually opens it.
func TestOpen_door(t *testing.T) {

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

	// Try to open the door
	cmd.Parse(actor, "open door")

	// Door should now be opened
	if !door.Opened() {
		t.Errorf("%s, %s: was not opened.",
			door.Parent(), attr.FindName(door.Parent()).Name("?"),
		)
	}

	locA.Free()
}
