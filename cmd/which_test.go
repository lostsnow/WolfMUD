// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd_test

import (
	"fmt"
	"testing"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/text"
)

func whichSetupWorld() (world attr.Things) {

	locA := attr.NewThing(
		attr.NewStart(),
		attr.NewName("Test room A"),
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
				attr.NewName("a small green ball"),
				attr.NewAlias("+SMALL", "+GREEN", "BALL"),
				attr.NewDescription("This is a small, green rubber ball."),
			),
			attr.NewThing(
				attr.NewName("a large green ball"),
				attr.NewAlias("+LARGE", "+GREEN", "BALL"),
				attr.NewDescription("This is a large, green rubber ball."),
			),
		),
	)

	return attr.Things{locA}
}

func TestWhich(t *testing.T) {

	world := whichSetupWorld()
	defer world.Free()

	actor := cmd.NewTestPlayer("an actor",
		attr.NewThing(
			attr.NewName("a token"),
			attr.NewAlias("+TEST", "TOKEN"),
			attr.NewDescription("This is a test token."),
		),
	)
	observer := cmd.NewTestPlayer("an observer")

	const (
		// For actor
		noFrog     = text.Bad + "You see no 'FROG' here.\n"
		smallGreen = text.Good + "You see a small green ball here.\n"
		largeGreen = text.Good + "You see a large green ball here.\n"
		fewerBalls = text.Bad + "You don't see that many 'BALL' here.\n"
		token      = text.Good + "You are carrying a token.\n"
		paintedA   = text.Good + "You see a large, painted letter 'A' here.\n"

		// For observer
		nothing = ""
		noting  = text.Reset + "\n" + text.Info +
			"The actor looks around taking note of various things.\n"
		notFound = text.Reset + "\n" + text.Info +
			"The actor looks around for something.\n"
	)

	for _, test := range []struct {
		cmd      string
		actor    string
		observer string
	}{
		// Tests to make sure items found from correct Inventories and colours are
		// set correctly for good/bad outcomes.
		{"", text.Info + "You look around for nothing in particular.\n", nothing},
		{"ball", smallGreen, noting},
		{"3rd ball", fewerBalls, notFound},
		{"frog", noFrog, notFound},
		{"small ball frog large ball", smallGreen + noFrog + largeGreen, noting},
		{"frog ball", noFrog + smallGreen, noting},
		{"ball frog", smallGreen + noFrog, noting},
		{"3rd ball token", fewerBalls + token, noting},
		{"token 3rd ball", token + fewerBalls, noting},
		{"token", token, noting},
		{"letter", paintedA, noting},
	} {
		t.Run(fmt.Sprintf("%s", test.cmd), func(t *testing.T) {
			cmd.Parse(actor, "which "+test.cmd)
			if have := actor.Messages(); have != test.actor {
				t.Errorf(
					"Actor for %+q:\nhave: %+q\nwant: %+q",
					"which "+test.cmd, have, test.actor,
				)
			}
			if have := observer.Messages(); have != test.observer {
				t.Errorf(
					"Observer for %+q:\nhave: %+q\nwant: %+q",
					"which "+test.cmd, have, test.observer,
				)
			}
		})
	}
}

func BenchmarkWhich(b *testing.B) {

	world := whichSetupWorld()
	defer world.Free()

	actor := cmd.NewTestPlayer("an actor")
	observer := cmd.NewTestPlayer("an observer")

	for _, test := range []string{
		"",         // Nothing
		"ball",     // 1st of several
		"all ball", // Several items
		"token",    // In actor's Inventory
		"frog",     // Not found
	} {
		c := "which " + test
		b.Run(fmt.Sprintf(test), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				cmd.Parse(actor, c)
				b.StopTimer()
				actor.Reset()
				observer.Reset()
				b.StartTimer()
			}
		})
	}
}
