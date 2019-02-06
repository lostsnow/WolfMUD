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
			"This is a room for testing. A large 'A' is painted on the wall.",
		),
		attr.NewInventory(
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
				attr.NewName("a small red ball"),
				attr.NewAlias("+SMALL", "+RED", "BALL"),
				attr.NewDescription("This is a small, red rubber ball."),
			),
			attr.NewThing(
				attr.NewName("a large red ball"),
				attr.NewAlias("+LARGE", "+RED", "BALL"),
				attr.NewDescription("This is a large, red rubber ball."),
			),
			attr.NewThing(
				attr.NewName("an apple"),
				attr.NewAlias("APPLE"),
				attr.NewDescription("This is an apple."),
			),
			attr.NewThing(
				attr.NewName("a piece of chalk"),
				attr.NewAlias("CHALK"),
				attr.NewDescription("This is a short stick of white chalk"),
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
		noBalls    = text.Bad + "You don't see that many 'BALL' here.\n"
		smallGreen = text.Good + "You see a small green ball here.\n"
		largeGreen = text.Good + "You see a large green ball here.\n"
		smallRed   = text.Good + "You see a small red ball here.\n"
		largeRed   = text.Good + "You see a large red ball here.\n"
		apple      = text.Good + "You see an apple here.\n"
		chalk      = text.Good + "You see a piece of chalk here.\n"
		token      = text.Good + "You are carrying a token.\n"

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
		{"", text.Good + "You look around for nothing in particular.\n", nothing},
		{"ball", smallGreen, noting},
		{"all ball", smallGreen + largeGreen + smallRed + largeRed, noting},
		{"green ball", smallGreen, noting},
		{"all green ball", smallGreen + largeGreen, noting},
		{"red ball", smallRed, noting},
		{"all red ball", smallRed + largeRed, noting},
		{"small ball", smallGreen, noting},
		{"all small ball", smallGreen + smallRed, noting},
		{"0 ball", noBalls, notFound},
		{"1 ball", smallGreen, noting},
		{"2 ball", smallGreen + largeGreen, noting},
		{"3 ball", smallGreen + largeGreen + smallRed, noting},
		{"4 ball", smallGreen + largeGreen + smallRed + largeRed, noting},
		{"5 ball", smallGreen + largeGreen + smallRed + largeRed, noting},
		{"0-0 ball", noBalls, notFound},
		{"0-1 ball", smallGreen, noting},
		{"0-2 ball", smallGreen + largeGreen, noting},
		{"0-3 ball", smallGreen + largeGreen + smallRed, noting},
		{"0-4 ball", smallGreen + largeGreen + smallRed + largeRed, noting},
		{"1-5 ball", smallGreen + largeGreen + smallRed + largeRed, noting},
		{"1-2 ball", smallGreen + largeGreen, noting},
		{"1-3 ball", smallGreen + largeGreen + smallRed, noting},
		{"1-4 ball", smallGreen + largeGreen + smallRed + largeRed, noting},
		{"1-5 ball", smallGreen + largeGreen + smallRed + largeRed, noting},
		{"2-1 ball", smallGreen + largeGreen, noting},
		{"2-2 ball", largeGreen, noting},
		{"2-3 ball", largeGreen + smallRed, noting},
		{"2-4 ball", largeGreen + smallRed + largeRed, noting},
		{"2-5 ball", largeGreen + smallRed + largeRed, noting},
		{"3-1 ball", smallGreen + largeGreen + smallRed, noting},
		{"3-2 ball", largeGreen + smallRed, noting},
		{"3-3 ball", smallRed, noting},
		{"3-4 ball", smallRed + largeRed, noting},
		{"3-5 ball", smallRed + largeRed, noting},
		{"4-1 ball", smallGreen + largeGreen + smallRed + largeRed, noting},
		{"4-2 ball", largeGreen + smallRed + largeRed, noting},
		{"4-3 ball", smallRed + largeRed, noting},
		{"4-4 ball", largeRed, noting},
		{"4-5 ball", largeRed, noting},
		{"5-5 ball", noBalls, notFound},
		{"5-6 ball", noBalls, notFound},
		{"0th ball", noBalls, notFound},
		{"1st ball", smallGreen, noting},
		{"2nd ball", largeGreen, noting},
		{"3rd ball", smallRed, noting},
		{"4th ball", largeRed, noting},
		{"5th ball", noBalls, notFound},
		{"frog", text.Bad + "You see no 'FROG' here.\n", notFound},
		{"blue frog", text.Bad + "You see no 'BLUE FROG' here.\n", notFound},
		{"green frog", text.Bad + "You see no 'GREEN FROG' here.\n", notFound},
		{"small frog", text.Bad + "You see no 'SMALL FROG' here.\n", notFound},
		{
			"red frog ball",
			text.Bad + "You see no 'RED FROG' here.\n" + smallGreen, noting,
		},
		{"apple ball chalk", apple + smallGreen + chalk, noting},
		{"apple 0th ball chalk", apple + noBalls + chalk, noting},
		{
			"apple all ball chalk",
			apple + smallGreen + largeGreen + smallRed + largeRed + chalk, noting,
		},
		{"token", token, noting},
		{"apple token chalk", apple + token + chalk, noting},

		// Should not find a qualifier as an alias
		// BUG (diddymus): test disabled until aliases updated to handle qualifiers
		// {"+test", "You see no '+TEST' here.", notFound},
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
		"",
		"apple",
		"ball",
		"token",
		"all ball",
		"all green ball",
		"all red ball",
		"all small ball",
		"all large ball",
		"apple ball chalk",
		"apple all ball chalk",
		"frog",
		"apple frog",
		"apple frog chalk",
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
