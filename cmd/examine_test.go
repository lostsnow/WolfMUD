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

// TestExamine_messages checks messages are output in the correct order with
// the correct color as well as being sent to the right players.
func TestExamine_messages(t *testing.T) {

	// Observer Reset+Info shorthand
	const ORI = text.Reset + "\n" + text.Info

	for _, test := range []struct {
		params   string
		actor    string
		observer string
	}{
		{
			"", // No item
			text.Info + "You examine this and that, find nothing special.\n", "",
		}, {
			"frog", // Invalid item
			text.Bad + "You see no 'FROG' to examine.\n", "",
		}, {
			"rock", // Single simple item at location
			text.Good + "You examine the rock." +
				text.Reset + "\nThis is a small rock.\n",
			ORI + "The actor studies a rock.\n",
		}, {
			"cup", // Examine empty container at location
			text.Good + "You examine the cup." +
				text.Reset + "\nThis is a cup. It is empty.\n",
			ORI + "The actor studies a cup.\n",
		}, {
			"box", // Examine container with single item at location
			text.Good + "You examine the box." +
				text.Reset + "\nThis is a box. It contains a small green ball.\n",
			ORI + "The actor studies a box.\n",
		}, {
			"bag", // Examine container with multile items at location
			text.Good + "You examine the bag." +
				text.Reset + "\nThis is a bag. It contains:\n" +
				"  a small green ball\n" +
				"  a small red ball\n",
			ORI + "The actor studies a bag.\n",
		}, {
			"token", // Single simple held item
			text.Good + "You examine the token." +
				text.Reset + "\nThis is a test token.\n",
			ORI + "The actor studies a token they are carrying.\n",
		}, {
			"mug", // Examine empty, held container
			text.Good + "You examine the mug." +
				text.Reset + "\nThis is a mug. It is empty.\n",
			ORI + "The actor studies a mug they are carrying.\n",
		}, {
			"pouch", // Examine held container with single item
			text.Good + "You examine the pouch." +
				text.Reset + "\nThis is a pouch. It contains a small green ball.\n",
			ORI + "The actor studies a pouch they are carrying.\n",
		}, {
			"bucket", // Examine held container with multiple items
			text.Good + "You examine the bucket." +
				text.Reset + "\nThis is a small, plastic bucket. It contains:\n" +
				"  some sand\n" +
				"  a pretty seashell\n",
			ORI + "The actor studies a bucket they are carrying.\n",
		}, {
			"stone", // Examine held item also at location - should pick location item
			text.Good + "You examine the stone." +
				text.Reset + "\nThis is a large stone.\n",
			ORI + "The actor studies a stone.\n",
		}, {
			"door", // Examine a closed door
			text.Good + "You examine the door." +
				text.Reset + "\nThis is a door. It is closed.\n",
			ORI + "The actor studies a door.\n",
		}, {
			"window", // Examine an open window
			text.Good + "You examine the window." +
				text.Reset + "\nThis is a window. It is open.\n",
			ORI + "The actor studies a window.\n",
		}, {
			"parchament", // Examine a vetoed item
			text.Bad + "The text on the ancient parchament swirls before you eyes.\n",
			"",
		},
	} {

		world := attr.NewThing(
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
					attr.NewDoor(attr.East, true, time.Second, 0),
					attr.NewNarrative(),
				),
				attr.NewThing(
					attr.NewName("a rock"),
					attr.NewAlias("ROCK"),
					attr.NewDescription("This is a small rock."),
				),
				attr.NewThing(
					attr.NewName("a stone"),
					attr.NewAlias("STONE"),
					attr.NewDescription("This is a large stone."),
				),
				attr.NewThing(
					attr.NewName("a cup"),
					attr.NewAlias("CUP"),
					attr.NewDescription("This is a cup."),
					attr.NewInventory(),
				),
				attr.NewThing(
					attr.NewName("a box"),
					attr.NewAlias("BOX"),
					attr.NewDescription("This is a box."),
					attr.NewInventory(
						attr.NewThing(
							attr.NewName("a small green ball"),
							attr.NewAlias("BALL"),
							attr.NewDescription("This is a small, green rubber ball."),
						),
					),
				),
				attr.NewThing(
					attr.NewName("a bag"),
					attr.NewAlias("BAG"),
					attr.NewDescription("This is a bag."),
					attr.NewInventory(
						attr.NewThing(
							attr.NewName("a small green ball"),
							attr.NewAlias("BALL"),
							attr.NewDescription("This is a small, green rubber ball."),
						),
						attr.NewThing(
							attr.NewName("a small red ball"),
							attr.NewAlias("BALL"),
							attr.NewDescription("This is a small, red rubber ball."),
						),
					),
				),
			),
		)

		actor := cmd.NewTestPlayer("an actor", "ACTOR",
			attr.NewThing(
				attr.NewName("a stone"),
				attr.NewAlias("STONE"),
				attr.NewDescription("This is a small stone."),
			),
			attr.NewThing(
				attr.NewName("a token"),
				attr.NewAlias("+TEST", "TOKEN"),
				attr.NewDescription("This is a test token."),
			),
			attr.NewThing(
				attr.NewName("a mug"),
				attr.NewAlias("MUG"),
				attr.NewDescription("This is a mug."),
				attr.NewInventory(),
			),
			attr.NewThing(
				attr.NewName("a pouch"),
				attr.NewAlias("POUCH"),
				attr.NewDescription("This is a pouch."),
				attr.NewInventory(
					attr.NewThing(
						attr.NewName("a small green ball"),
						attr.NewAlias("BALL"),
						attr.NewDescription("This is a small, green rubber ball."),
					),
				),
			),
			attr.NewThing(
				attr.NewName("an ancient parchament"),
				attr.NewAlias("PARCHAMENT"),
				attr.NewDescription("This is an ancient parchament."),
				attr.NewVetoes(
					attr.NewVeto("EXAM", "The text on the ancient parchament swirls before you eyes."),
				),
			),
			attr.NewThing(
				attr.NewName("a bucket"),
				attr.NewAlias("BUCKET"),
				attr.NewDescription("This is a small, plastic bucket."),
				attr.NewInventory(
					attr.NewThing(
						attr.NewName("some sand"),
						attr.NewAlias("SAND"),
						attr.NewDescription("This is a small amount of sand."),
					),
					attr.NewThing(
						attr.NewName("a pretty seashell"),
						attr.NewAlias("SEASHELL"),
						attr.NewDescription("This is a pretty seashell."),
					),
				),
			),
		)

		observer := cmd.NewTestPlayer("observer", "OBSERVER")

		c := "examine " + test.params
		t.Run(c, func(t *testing.T) {
			cmd.Parse(actor, c)
			if have := actor.Messages(); have != test.actor {
				t.Errorf("Actor for %+q:\nhave: %+q\nwant: %+q", c, have, test.actor)
			}
			if have := observer.Messages(); have != test.observer {
				t.Errorf("Observer for %+q:\nhave: %+q\nwant: %+q", c, have, test.observer)
			}
		})

		world.Free()
	}
}

// TestExamine_player checks that when studying another player that the player
// becomes the participant of the EXAMINE command and that the participants
// inventory is not revealed.
func TestExamine_player(t *testing.T) {

	// Observer Reset+Info shorthand
	const ORI = text.Reset + "\n" + text.Info

	for _, test := range []struct {
		params      string
		actor       string
		observer    string
		participant string
	}{
		{
			"participant",
			text.Good + "You examine the participant." +
				text.Reset + "\nThis is a test player.\n",
			ORI + "The actor studies a participant.\n",
			ORI + "The actor studies you.\n",
		},
	} {

		world := attr.NewThing(
			attr.NewStart(),
			attr.NewName("Test room A"),
			attr.NewAlias("ROOM_A"),
			attr.NewDescription(
				"This is a room for testing.",
			),
			attr.NewInventory(),
			attr.NewExits(),
		)

		actor := cmd.NewTestPlayer("an actor", "ACTOR",
			attr.NewThing(
				attr.NewName("a token"),
				attr.NewAlias("+TEST", "TOKEN"),
				attr.NewDescription("This is a test token."),
			),
		)

		observer := cmd.NewTestPlayer("an observer", "OBSERVER")

		participant := cmd.NewTestPlayer("a participant", "PARTICIPANT",
			attr.NewThing(
				attr.NewName("a mug"),
				attr.NewAlias("MUG"),
				attr.NewDescription("This is a mug."),
				attr.NewInventory(),
			),
		)

		c := "examine " + test.params
		t.Run(c, func(t *testing.T) {
			cmd.Parse(actor, c)
			if have := actor.Messages(); have != test.actor {
				t.Errorf("Actor for %+q:\nhave: %+q\nwant: %+q", c, have, test.actor)
			}
			if have := observer.Messages(); have != test.observer {
				t.Errorf("Observer for %+q:\nhave: %+q\nwant: %+q", c, have, test.observer)
			}
			if have := participant.Messages(); have != test.participant {
				t.Errorf("Participant for %+q:\nhave: %+q\nwant: %+q", c, have, test.participant)
			}
		})

		world.Free()
	}
}
