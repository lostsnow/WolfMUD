// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd_test

import (
	"strings"
	"testing"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/has"
)

type test struct {
	words     string
	results   []cmd.Result
	remaining string
}

func match(words string) *test {
	return &test{words, []cmd.Result{}, ""}
}

func (t *test) found(things ...has.Thing) *test {
	for _, thing := range things {
		t.results = append(t.results, cmd.Result{thing, "", ""})
	}
	return t
}

func (t *test) unknown(unknowns ...string) *test {
	for _, u := range unknowns {
		t.results = append(t.results, cmd.Result{nil, strings.ToUpper(u), ""})
	}
	return t
}

func (t *test) notEnough(notEnoughs ...string) *test {
	for _, ne := range notEnoughs {
		t.results = append(t.results, cmd.Result{nil, "", strings.ToUpper(ne)})
	}
	return t
}

func (t *test) remains(words string) *test {
	t.remaining = strings.ToUpper(words)
	return t
}

func list(r []cmd.Result) string {
	var list []string
	for _, r := range r {
		switch {
		case r.Unknown != "":
			list = append(list, "UNKNOWN:"+r.Unknown)
		case r.NotEnough != "":
			list = append(list, "NOT_ENOUGH:"+r.NotEnough)
		default:
			list = append(list, attr.FindName(r).Name("?"))
		}
	}
	return strings.Join(list, ", ")
}

func item(name string, aliases ...string) has.Thing {
	return attr.NewThing(attr.NewName(name), attr.NewAlias(aliases...))
}

func TestMatchAll(t *testing.T) {

	smallGreen := item("a small green ball", "+SMALL", "+GREEN", "BALL")
	largeGreen := item("a large green ball", "+LARGE", "+GREEN", "BALL")
	smallRed := item("a small red ball", "+SMALL", "+RED", "BALL")
	largeRed := item("a large red ball", "+LARGE", "+RED", "BALL")
	shortSword :=
		item("a wooden shortsword", "+WOODEN", "+SHORT:SWORD", "SHORTSWORD")
	longSword :=
		item("a wooden longsword", "+WOODEN", "+LONG:SWORD", "LONGSWORD")
	apple := item("an apple", "APPLE")
	chalk := item("some chalk", "CHALK")
	token := item("a token", "+TEST", "TOKEN")

	items := []has.Thing{
		smallGreen, largeGreen, smallRed, largeRed,
		shortSword, longSword, apple, chalk, token,
	}

	for _, test := range []*test{
		match(""),

		match("ball").found(smallGreen),
		match("all ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("green ball").found(smallGreen),
		match("all green ball").found(smallGreen, largeGreen),
		match("red ball").found(smallRed),
		match("all red ball").found(smallRed, largeRed),
		match("small ball").found(smallGreen),
		match("all small ball").found(smallGreen, smallRed),
		match("0 ball").notEnough("ball"),
		match("1 ball").found(smallGreen),
		match("2 ball").found(smallGreen, largeGreen),
		match("3 ball").found(smallGreen, largeGreen, smallRed),
		match("4 ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("5 ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("0-0 ball").notEnough("ball"),
		match("0-1 ball").found(smallGreen),
		match("0-2 ball").found(smallGreen, largeGreen),
		match("0-3 ball").found(smallGreen, largeGreen, smallRed),
		match("0-4 ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("0-5 ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("1-0 ball").found(smallGreen),
		match("1-1 ball").found(smallGreen),
		match("1-2 ball").found(smallGreen, largeGreen),
		match("1-3 ball").found(smallGreen, largeGreen, smallRed),
		match("1-4 ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("1-5 ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("2-0 ball").found(smallGreen, largeGreen),
		match("2-1 ball").found(smallGreen, largeGreen),
		match("2-2 ball").found(largeGreen),
		match("2-3 ball").found(largeGreen, smallRed),
		match("2-4 ball").found(largeGreen, smallRed, largeRed),
		match("2-5 ball").found(largeGreen, smallRed, largeRed),
		match("3-0 ball").found(smallGreen, largeGreen, smallRed),
		match("3-1 ball").found(smallGreen, largeGreen, smallRed),
		match("3-2 ball").found(largeGreen, smallRed),
		match("3-3 ball").found(smallRed),
		match("3-4 ball").found(smallRed, largeRed),
		match("3-5 ball").found(smallRed, largeRed),
		match("4-0 ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("4-1 ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("4-2 ball").found(largeGreen, smallRed, largeRed),
		match("4-3 ball").found(smallRed, largeRed),
		match("4-4 ball").found(largeRed),
		match("4-5 ball").found(largeRed),
		match("5-0 ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("5-1 ball").found(smallGreen, largeGreen, smallRed, largeRed),
		match("5-2 ball").found(largeGreen, smallRed, largeRed),
		match("5-3 ball").found(smallRed, largeRed),
		match("5-4 ball").found(largeRed),
		match("5-5 ball").notEnough("ball"),
		match("0th ball").notEnough("ball"),
		match("1st ball").found(smallGreen),
		match("2nd ball").found(largeGreen),
		match("3rd ball").found(smallRed),
		match("4th ball").found(largeRed),
		match("5th ball").notEnough("ball"),

		match("all small").unknown("all small"),
		match("frog").unknown("frog"),
		match("blue frog").unknown("blue frog"),
		match("green frog").unknown("green frog"),
		match("small frog").unknown("small frog"),
		match("red frog ball").unknown("red frog").found(smallGreen),
		match("apple ball chalk").found(apple, smallGreen, chalk),
		match("apple 0th ball chalk").found(apple).notEnough("ball").found(chalk),
		match("apple all ball chalk").
			found(apple, smallGreen, largeGreen, smallRed, largeRed, chalk),
		match("token").found(token),
		match("3rd token").notEnough("token"),
		match("apple token chalk").found(apple, token, chalk),

		// Tests for unique results and results with overlapping ranges
		match("token token").found(token),
		match("token apple token").found(apple, token),
		match("all ball all ball").
			found(smallGreen, largeGreen, smallRed, largeRed),
		match("1-4 ball 1-4 ball").
			found(smallGreen, largeGreen, smallRed, largeRed),
		match("1-3 ball 2-4 ball").
			found(smallGreen, largeGreen, smallRed, largeRed),
		match("all ball 1-4 ball").
			found(smallGreen, largeGreen, smallRed, largeRed),
		match("all ball 1-3 ball 2-4 ball").
			found(smallGreen, largeGreen, smallRed, largeRed),

		// Should not find a qualifier or bound qualifier as an alias
		match("+test").unknown("+test"),
		match("+short").unknown("+short"),
		match("+short:sword").unknown("+short:sword"),
		match("short:sword").unknown("short:sword"),

		// Tests for qualifier bound to aliases
		match("all shortsword").found(shortSword),
		match("all longsword").found(longSword),
		match("all short sword").found(shortSword),
		match("all long sword").found(longSword),
		match("all sword").found(shortSword, longSword),
		match("all wooden sword").found(shortSword, longSword),
		match("all wooden shortsword").found(shortSword),
		match("all wooden longsword").found(longSword),
		match("all wooden short sword").found(shortSword),
		match("all wooden long sword").found(longSword),
		match("all short wooden sword").found(shortSword),
		match("all long wooden sword").found(longSword),
		match("short").unknown("short"),
		match("long").unknown("long"),
		match("wooden").unknown("wooden"),
		match("wooden short").unknown("wooden short"),
		match("wooden long").unknown("wooden long"),
		match("short shortsword").unknown("short").found(shortSword),
		match("long longsword").unknown("long").found(longSword),

		// These four tests detect errors when found, not enough and unknown set in
		// same result, which should not happen.
		match("chalk 2nd apple ball").
			found(chalk).notEnough("apple").found(smallGreen),
		match("chalk ball 2nd apple").
			found(chalk, smallGreen).notEnough("apple"),
		match("chalk 2nd apple frog").
			found(chalk).notEnough("apple").unknown("frog"),
		match("chalk frog 2nd apple").
			found(chalk).unknown("frog").notEnough("apple"),
	} {
		t.Run(test.words, func(t *testing.T) {
			words := strings.Fields(strings.ToUpper(test.words))
			have := cmd.MatchAll(words, items)
			haveList := list(have)
			wantList := list(test.results)
			if haveList != wantList {
				t.Errorf("\nhave: %s\nwant: %s", haveList, wantList)
			}
		})
	}
}

func BenchmarkMatchAll(b *testing.B) {

	items := []has.Thing{
		item("a small green ball", "+SMALL", "+GREEN", "BALL"),
		item("a large green ball", "+LARGE", "+GREEN", "BALL"),
		item("a small red ball", "+SMALL", "+RED", "BALL"),
		item("a large red ball", "+LARGE", "+RED", "BALL"),
		item("a wooden shortsword", "+WOODEN", "+SHORT:SWORD", "SHORTSWORD"),
		item("a wooden longsword", "+WOODEN", "+LONG:SWORD", "LONGSWORD"),
		item("an apple", "APPLE"),
		item("some chalk", "CHALK"),
		item("a token", "+TEST", "TOKEN"),
	}

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
		"a b c d e f g h i j k l m n o p q r s t u v w x z y",
	} {
		b.Run(test, func(b *testing.B) {
			words := strings.Fields(strings.ToUpper(test))
			for i := 0; i < b.N; i++ {
				_ = cmd.MatchAll(words, items)
			}
		})
	}
}

func TestMatch(t *testing.T) {

	smallGreen := item("a small green ball", "+SMALL", "+GREEN", "BALL")
	largeGreen := item("a large green ball", "+LARGE", "+GREEN", "BALL")
	token := item("a token", "+TEST", "TOKEN")

	items := []has.Thing{smallGreen, largeGreen, token}

	for _, test := range []*test{
		match(""),
		match("ball").found(smallGreen),
		match("frog").unknown("frog"),
		match("green frog").unknown("frog").remains("green"),
		match("ball token").found(token).remains("ball"),
		match("small ball token").found(token).remains("small ball"),
		match("green ball token").found(token).remains("green ball"),
		match("small green ball token").found(token).remains("small green ball"),
		match("token ball").found(smallGreen).remains("token"),
		match("token frog").unknown("frog").remains("token"),
		match("token small ball").found(smallGreen).remains("token"),
		match("token green ball").found(smallGreen).remains("token"),
		match("token small green ball").found(smallGreen).remains("token"),
		match("token green frog ball").found(smallGreen).remains("token green frog"),
		match("token all ball").found(smallGreen, largeGreen).remains("token"),
	} {
		t.Run(test.words, func(t *testing.T) {
			words := strings.Fields(strings.ToUpper(test.words))
			matches, words := cmd.Match(words, items)
			haveList := list(matches)
			haveWords := strings.Join(words, " ")
			wantList := list(test.results)
			if haveList != wantList || haveWords != test.remaining {
				t.Errorf("\nhave: %s (%s)\nwant: %s (%s)",
					haveList, haveWords, wantList, test.remaining,
				)
			}
		})
	}
}

func BenchmarkMatch(b *testing.B) {

	items := []has.Thing{
		item("a small green ball", "+SMALL", "+GREEN", "BALL"),
		item("a large green ball", "+LARGE", "+GREEN", "BALL"),
		item("a small red ball", "+SMALL", "+RED", "BALL"),
		item("a large red ball", "+LARGE", "+RED", "BALL"),
		item("a wooden shortsword", "+WOODEN", "+SHORT:SWORD", "SHORTSWORD"),
		item("a wooden longsword", "+WOODEN", "+LONG:SWORD", "LONGSWORD"),
		item("an apple", "APPLE"),
		item("some chalk", "CHALK"),
		item("a token", "+TEST", "TOKEN"),
	}

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
		"a b c d e f g h i j k l m n o p q r s t u v w x z y",
	} {
		b.Run(test, func(b *testing.B) {
			words := strings.Fields(strings.ToUpper(test))
			for i := 0; i < b.N; i++ {
				_, _ = cmd.Match(words, items)
			}
		})
	}
}
