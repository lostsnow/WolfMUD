// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd_test

import (
	"fmt"
	"testing"
	"time"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/text"
)

// TestDrop_messages checks messages are output in the correct order with the
// correct color as well as being sent to the right players.
func TestDrop_messages(t *testing.T) {

	const P = "\n" + text.Prompt // Prompt (StyleNone) shorthand

	for _, test := range []struct {
		cmd      string
		actor    string
		observer string
	}{
		// Drop with no item specified
		{"", text.Info + "You go to drop... something?" + P, ""},

		// Try to drop a single item
		{
			"ball",
			text.Good + "You drop the small green ball." + P,
			"\n" + text.Info + "The actor drops a small green ball." + P,
		},

		// Duplicate normal drop - if world not cleaned after each test it will fail
		{
			"ball",
			text.Good + "You drop the small green ball." + P,
			"\n" + text.Info + "The actor drops a small green ball." + P,
		},

		// Try to drop multiple items
		{
			"small ball large ball",
			text.Good + "You drop the small green ball.\n" +
				text.Good + "You drop the large green ball." + P,
			"\n" +
				text.Info + "The actor drops a small green ball.\n" +
				text.Info + "The actor drops a large green ball." + P,
		},

		// Try to drop multiple items with one being invalid - checks colour change
		{
			"small ball frog large ball",
			text.Good + "You drop the small green ball.\n" +
				text.Bad + "You have no 'FROG' to drop.\n" +
				text.Good + "You drop the large green ball." + P,
			"\n" +
				text.Info + "The actor drops a small green ball.\n" +
				text.Info + "The actor drops a large green ball." + P,
		},

		// Try to drop an invalid item
		{"frog", text.Bad + "You have no 'FROG' to drop." + P, ""},

		// Try to drop too many of an item
		{"3rd ball", text.Bad + "You don't have that many 'BALL' to drop." + P, ""},

		// Try to drop a non-narrative with an overriding veto message
		{"something sticky", text.Bad + "You can't let go of something sticky." + P, ""},
	} {

		world := attr.Things{
			attr.NewThing(
				attr.NewStart(),
				attr.NewName("Test room A"),
				attr.NewAlias("ROOM_A"),
				attr.NewDescription(
					"This is a room for testing. A large letter 'A' is painted on the wall.",
				),
				attr.NewInventory(),
			),
		}

		actor := cmd.NewTestPlayer("an actor", "ACTOR",
			attr.NewThing(
				attr.NewName("a token"),
				attr.NewAlias("+TEST", "TOKEN"),
				attr.NewDescription("This is a test token."),
			),
			attr.NewThing(
				attr.NewName("a small green ball"),
				attr.NewAlias("+SMALL", "+GREEN", "BALL"),
				attr.NewDescription("This is a small, green rubber ball."),
			),
			attr.NewThing(
				attr.NewName("a large green ball"),
				attr.NewAlias("+LARGE", "+GREEN", "BALL"),
				attr.NewDescription("This is a large, green rubber ball."),
			),
			attr.NewThing(
				attr.NewName("something sticky"),
				attr.NewAlias("+SOMETHING:STICKY"),
				attr.NewDescription("This is something sticky"),
				attr.NewVetoes(
					attr.NewVeto("DROP", "You can't let go of something sticky."),
				),
			),
		)
		observer := cmd.NewTestPlayer("an observer", "OBSERVER")

		t.Run(fmt.Sprintf("%s", test.cmd), func(t *testing.T) {
			cmd.Parse(actor, "drop "+test.cmd)
			if have := actor.Messages(); have != test.actor {
				t.Errorf(
					"Actor for %+q:\nhave: %+q\nwant: %+q",
					"drop "+test.cmd, have, test.actor,
				)
			}
			if have := observer.Messages(); have != test.observer {
				t.Errorf(
					"Observer for %+q:\nhave: %+q\nwant: %+q",
					"drop "+test.cmd, have, test.observer,
				)
			}

			world.Free()
		})
	}
}

// Make sure we handle the actor not having an Inventory to drop things from.
func TestDrop_noInventory(t *testing.T) {

	locA := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		attr.NewInventory(),
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR")

	// Remove and free player's inventory
	inv := attr.FindInventory(actor)
	actor.Remove(inv)
	inv.Free()

	// Try and drop an item when player has no inventory
	cmd.Parse(actor, "drop ball")
	have := actor.Messages()
	want := text.Bad + "You don't have anything to drop.\n" + text.Prompt
	if have != want {
		t.Errorf(
			"Actor for %+q:\nhave: %+q\nwant: %+q",
			"drop ball", have, want,
		)
	}

	locA.Free()
}

// TestDrop_inventory check to make sure dropping an item moves it from the
// actor's Inventory.
func TestDrop_inventory(t *testing.T) {

	// Piece together world keeping references

	inv := attr.NewInventory()

	locA := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		inv,
	)

	ball := attr.NewThing(
		attr.NewName("a small green ball"),
		attr.NewAlias("+SMALL", "+GREEN", "BALL"),
	)
	uid := ball.UID()

	actor := cmd.NewTestPlayer("an actor", "ACTOR", ball)

	// Try and drop small ball
	cmd.Parse(actor, "drop small green ball")

	// Small ball should be found in location Inventory
	if inv.Search(uid) == nil {
		t.Errorf("%s, %s: not in location inventory.",
			ball, attr.FindName(ball).Name("?"),
		)
	}

	// Small ball should now not be in Player Inventory
	if attr.FindInventory(actor).Search(uid) != nil {
		t.Errorf("%s, %s: in player inventory.",
			ball, attr.FindName(ball).Name("?"),
		)
	}

	locA.Free()
}

// TestDrop_events tests to make sure action and cleanup events are enabled
// correctly when we drop an item.
func TestDrop_events(t *testing.T) {

	locA := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		attr.NewInventory(),
	)

	cleanup := attr.NewCleanup(time.Hour, 0)
	action := attr.NewAction(time.Hour, 0)

	ball := attr.NewThing(
		attr.NewName("a small green ball"),
		attr.NewAlias("+SMALL", "+GREEN", "BALL"),
		cleanup,
		action,
		attr.NewOnAction([]string{"The ball moves..."}),
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR", ball)

	// Check clean up and action events inactive
	checkEvent(t, cleanup, inactive)
	checkEvent(t, action, inactive)

	// Drop ball, should start cleanup and action events
	cmd.Parse(actor, "drop small green ball")
	checkEvent(t, cleanup, active)
	checkEvent(t, action, active)

	locA.Free()
}
