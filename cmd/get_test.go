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

// getSetupWorld creates a simple test world with some items.
func getSetupWorld() (world attr.Things) {

	locA := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		attr.NewAlias("ROOM_A"),
		attr.NewDescription(
			"This is a room for testing. A large letter 'A' is painted on the wall.",
		),
		attr.NewInventory(
			attr.NewThing(
				attr.NewName("a large, painted letter 'A'"),
				attr.NewAlias("+PAINTED", "+LARGE", "LETTER"),
				attr.NewDescription("This is a large, painted, capital letter 'A'."),
				attr.NewNarrative(),
			),
			attr.NewThing(
				attr.NewName("a wall"),
				attr.NewAlias("WALL"),
				attr.NewDescription("This is a brick wall holding up the ceiling."),
				attr.NewVetoes(
					attr.NewVeto("GET", "If you take the wall the ceiling will crush you!"),
				),
				attr.NewNarrative(),
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
				attr.NewName("some water"),
				attr.NewAlias("WATER"),
				attr.NewDescription("This is a small pool of water."),
				attr.NewVetoes(
					attr.NewVeto("GET", "The water runs through your fingers."),
				),
			),
		),
	)

	return attr.Things{locA}
}

// TestGet_messages checks messages are output in the correct order with the
// correct color as well as being sent to the right players.
func TestGet_messages(t *testing.T) {

	for _, test := range []struct {
		cmd      string
		actor    string
		observer string
	}{
		// Get with no item specified
		{"", text.Info + "You go to get... something?\n", ""},

		// Try to get a single item
		{
			"ball",
			text.Good + "You get a small green ball.\n",
			text.Reset + "\n" + text.Info + "You see the actor get a small green ball.\n",
		},

		// Duplicate normal get - if world not cleaned after each test it will fail
		{
			"ball",
			text.Good + "You get a small green ball.\n",
			text.Reset + "\n" + text.Info + "You see the actor get a small green ball.\n",
		},

		// Try to get multiple items
		{
			"small ball large ball",
			text.Good + "You get a small green ball.\n" +
				text.Good + "You get a large green ball.\n",
			text.Reset + "\n" +
				text.Info + "You see the actor get a small green ball.\n" +
				text.Info + "You see the actor get a large green ball.\n",
		},

		// Try to get multiple items with one being invalid - checks colour change
		{
			"small ball frog large ball",
			text.Good + "You get a small green ball.\n" +
				text.Bad + "You see no 'FROG' to get.\n" +
				text.Good + "You get a large green ball.\n",
			text.Reset + "\n" +
				text.Info + "You see the actor get a small green ball.\n" +
				text.Info + "You see the actor get a large green ball.\n",
		},

		// Try to get multiple items with one being a narrative - checks colour change
		{
			"small ball wall large ball",
			text.Good + "You get a small green ball.\n" +
				text.Bad + "If you take the wall the ceiling will crush you!\n" +
				text.Good + "You get a large green ball.\n",
			text.Reset + "\n" +
				text.Info + "You see the actor get a small green ball.\n" +
				text.Info + "You see the actor get a large green ball.\n",
		},

		// Try to get an invalid item
		{"frog", text.Bad + "You see no 'FROG' to get.\n", ""},

		// Try to get too many of an item
		{"3rd ball", text.Bad + "You don't see that many 'BALL' to get.\n", ""},

		// Try to get a narrative
		{
			"letter",
			text.Bad + "For some reason you cannot get a large, painted letter 'A'.\n",
			"",
		},

		// Try to get a narrative with an overriding veto message
		{"wall", text.Bad + "If you take the wall the ceiling will crush you!\n", ""},

		// Try to get a non-narrative with an overriding veto message
		{"water", text.Bad + "The water runs through your fingers.\n", ""},

		// Try to get self
		{"actor", text.Info + "Trying to pick youreself up by your bootlaces?\n", ""},

		// Try to another player
		{"observer", text.Bad + "The observer does not want to be picked up!\n", ""},

		// Try to get an item we are already carrying
		{"token", text.Bad + "You see no 'TOKEN' to get.\n", ""},
	} {

		world := getSetupWorld()

		actor := cmd.NewTestPlayer("an actor", "ACTOR",
			attr.NewThing(
				attr.NewName("a token"),
				attr.NewAlias("+TEST", "TOKEN"),
				attr.NewDescription("This is a test token."),
			),
		)
		observer := cmd.NewTestPlayer("an observer", "OBSERVER")

		t.Run(fmt.Sprintf("%s", test.cmd), func(t *testing.T) {
			cmd.Parse(actor, "get "+test.cmd)
			if have := actor.Messages(); have != test.actor {
				t.Errorf(
					"Actor for %+q:\nhave: %+q\nwant: %+q",
					"get "+test.cmd, have, test.actor,
				)
			}
			if have := observer.Messages(); have != test.observer {
				t.Errorf(
					"Observer for %+q:\nhave: %+q\nwant: %+q",
					"get "+test.cmd, have, test.observer,
				)
			}

			world.Free()
		})
	}
}

// Make sure we handle the actor not having an Inventory to put things in.
func TestGet_noInventory(t *testing.T) {

	world := getSetupWorld()
	actor := cmd.NewTestPlayer("an actor", "ACTOR")

	// Remove and free player's inventory
	inv := attr.FindInventory(actor)
	actor.Remove(inv)
	inv.Free()

	// Try and get an item when player has no inventory
	cmd.Parse(actor, "get ball")
	have := actor.Messages()
	want := text.Bad + "You can't carry anything!\n"
	if have != want {
		t.Errorf(
			"Actor for %+q:\nhave: %+q\nwant: %+q",
			"get ball", have, want,
		)
	}

	world.Free()
}

// TestGet_inventory check to make sure picking up an item moves it into the
// actor's Inventory.
func TestGet_inventory(t *testing.T) {

	// Piece together world keeping references

	ball := attr.NewThing(
		attr.NewName("a small green ball"),
		attr.NewAlias("+SMALL", "+GREEN", "BALL"),
	)
	uid := ball.UID()

	inv := attr.NewInventory(ball)

	locA := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		inv,
	)

	actor := cmd.NewTestPlayer("an actor", "ACTOR")

	// Try and get small ball
	cmd.Parse(actor, "get small green ball")

	// Small ball should not be found in location Inventory
	if inv.Search(uid) != nil {
		t.Errorf("%s, %s: still in location inventory.",
			ball, attr.FindName(ball).Name("?"),
		)
	}

	// Small ball should now be in Player Inventory
	if attr.FindInventory(actor).Search(uid) == nil {
		t.Errorf("%s, %s: not in player inventory.",
			ball, attr.FindName(ball).Name("?"),
		)
	}

	locA.Free()
}

// TestGet_spawnable checks that when a spawnable item is picked up we get a
// copy of the item and not the original.
func TestGet_spawnable(t *testing.T) {

	// Piece together world keeping references

	ball := attr.NewThing(
		attr.NewName("a small green ball"),
		attr.NewAlias("+SMALL", "+GREEN", "BALL"),
		attr.NewReset(time.Hour, 0, true),
	)
	uid := ball.UID()

	inv := attr.NewInventory(ball)

	locA := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		inv,
	)

	// Set origins so events work - usually done by the zone loader.
	locA.SetOrigins()

	actor := cmd.NewTestPlayer("an actor", "ACTOR")

	// Try and get small ball
	cmd.Parse(actor, "get small green ball")

	// Small ball should not be found in location Inventory
	if inv.Search(uid) != nil {
		t.Errorf("original ball still in location inventory.")
	}

	// Small ball should be disabled waiting for reset
	found := false
	for _, t := range inv.Disabled() {
		if t.UID() == uid {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("original ball not disabled at location.")
	}

	// Copy of small ball should now be in Player Inventory with different UID
	copy := attr.FindInventory(actor).Search("BALL")
	if copy == nil {
		t.Errorf("no ball found in player inventory.")
	}
	if copy.UID() == uid {
		t.Errorf("original ball in player inventory - should be a copy.")
	}

	locA.Free()
}

// Constants for event status
const (
	inactive = false
	active   = true
)

// checkEvent is a helper to test if an event is in the expected state.
func checkEvent(t *testing.T, p interface{ Pending() bool }, state bool) {
	t.Helper()
	have, want := "inactive", "active"
	if !state {
		have, want = want, have
	}
	if state != p.Pending() {
		t.Errorf("ball: %T event, have: %s, want: %s", p, have, want)
	}
}

// TestGet_events tests to make sure action and cleanup events are cancelled
// correctly when we get an item.
func TestGet_events(t *testing.T) {

	// Piece together world keeping references

	cleanup := attr.NewCleanup(time.Hour, 0)
	action := attr.NewAction(time.Hour, 0)

	ball := attr.NewThing(
		attr.NewName("a small green ball"),
		attr.NewAlias("+SMALL", "+GREEN", "BALL"),
		cleanup,
		action,
		attr.NewOnAction([]string{"The ball moves..."}),
	)

	inv := attr.NewInventory(ball)

	locA := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
		inv,
	)

	// Set origins so events work and start Action event - usually done by the
	// zone loader.
	locA.SetOrigins()
	action.Action()

	actor := cmd.NewTestPlayer("an actor", "ACTOR")

	// Start a clean up of the ball
	cleanup.Cleanup()

	// Check clean up and action events active
	checkEvent(t, cleanup, active)
	checkEvent(t, action, active)

	// Get ball, should stop cleanup and action events
	cmd.Parse(actor, "get small green ball")
	checkEvent(t, cleanup, inactive)
	checkEvent(t, action, inactive)

	locA.Free()
}
