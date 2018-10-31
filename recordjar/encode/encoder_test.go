// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package encode_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	. "code.wolfmud.org/WolfMUD.git/recordjar/encode"
)

func TestString(t *testing.T) {
	for _, test := range []struct {
		data string
		want string
	}{
		{"", ""},
		{" ", ""},
		{"\t", ""},
		{"all lowercase", "all lowercase"},
		{"ALL UPPERCASE", "ALL UPPERCASE"},
		{" Leading Space", "Leading Space"},
		{"Trailing Space ", "Trailing Space"},
		{" Both Space ", "Both Space"},
		{"\tLeading Tab", "Leading Tab"},
		{"Trailing Tab\t", "Trailing Tab"},
		{"\tBoth Tab\t", "Both Tab"},
		{"\u2007Unicode", "Unicode"},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := String(test.data)

			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

func BenchmarkString(b *testing.B) {
	for _, test := range []struct {
		name    string
		keyword string
	}{
		{"plain", "some text"},
		{"leading-space", " some text"},
		{"trailing-space", " some text "},
		{"both-space", " some text "},
		{"leading-tab", " some text"},
		{"trailing-tab", " some text "},
		{"both-tab", " some text "},
		{"long-both-space", " the quick brown fox jumps over the lazy dog "},
		{"figure-space", "\u2007Figure space"},
	} {
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = String(test.keyword)
			}
		})
	}
}

func TestKeyword(t *testing.T) {
	for _, test := range []struct {
		data string
		want string
	}{
		{"", ""},
		{" ", ""},
		{"\t", ""},
		{"keyword", "KEYWORD"},
		{" keyword", "KEYWORD"},
		{"keyword ", "KEYWORD"},
		{" keyword ", "KEYWORD"},
		{"\tkeyword", "KEYWORD"},
		{"keyword\t", "KEYWORD"},
		{"\tkeyword\t", "KEYWORD"},
		{"keyword\n", "KEYWORD"},
		{"\u2007keyword", "KEYWORD"},
		{"key word", "KEYWORD"},
		{"key\tword", "KEYWORD"},
		{"key\u2007word", "KEYWORD"},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := Keyword(test.data)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

func BenchmarkKeyword(b *testing.B) {
	for _, test := range []struct {
		name    string
		keyword string
	}{
		{"lower", "keyword"},
		{"upper", "KEYWORD"},
		{"mixed", "KeYwOrD"},
		{"split", "key word"},
		{"trim+lower", " keyword "},
		{"trim+upper", " KEYWORD "},
	} {
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Keyword(test.keyword)
			}
		})
	}
}

func TestKeywordList(t *testing.T) {
	for _, test := range []struct {
		data []string
		want string
	}{
		{[]string{}, ""},
		{[]string{""}, ""},
		{[]string{" "}, ""},
		{[]string{"\t"}, ""},
		{[]string{"", ""}, ""},
		{[]string{" ", " "}, ""},
		{[]string{"\t", "\t"}, ""},
		{[]string{"a", "keyword", "test"}, "A KEYWORD TEST"},
		{[]string{" a", "keyword ", " test "}, "A KEYWORD TEST"},
		{[]string{"\ta", "keyword\t", "\ttest\t"}, "A KEYWORD TEST"},
		{[]string{"key word"}, "KEYWORD"},
		{[]string{"key\tword"}, "KEYWORD"},
		{[]string{"key\u2007word"}, "KEYWORD"},
		{[]string{"z", "y", "x"}, "X Y Z"},
		{[]string{"ABC", "ABC", "XYZ", "XYZ"}, "ABC XYZ"},
		{[]string{"ABC", "abc", "XYZ", "xyz"}, "ABC XYZ"},
		{[]string{"abc", "ABC", "xyz", "XYZ"}, "ABC XYZ"},
		{[]string{"ABC", "XYZ", "ABC", "XYZ"}, "ABC XYZ"},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := KeywordList(test.data)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

func BenchmarkKeywordList(b *testing.B) {
	for _, test := range []struct {
		name     string
		keywords []string
	}{
		{"1x1", []string{"a"}},
		{"3x1", []string{"c", "b", "a"}},
		{"3x3", []string{"ccc", "bbb", "aaa"}},
		{"3x3Dup1", []string{"ABC", "ABC", "XYZ"}},
		{"3x3Dup2", []string{"ABC", "XYZ", "XYZ"}},
		{"3x3Dup3", []string{"ABC", "ABC", "ABC"}},
		{"6x1", []string{"f", "e", "d", "c", "b", "a"}},
		{"6x3", []string{"fff", "eee", "ddd", "ccc", "bbb", "aaa"}},
		{"6x10", []string{
			"ffffffffff", "eeeeeeeeee", "dddddddddd",
			"cccccccccc", "bbbbbbbbbb", "aaaaaaaaaa",
		}},
		{"Inventory-lower", []string{"l2n1", "l2n2", "l2n3", "m4"}},
		{"Inventory-upper", []string{"L2N1", "L2N2", "L2N3", "M4"}},
	} {
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = KeywordList(test.keywords)
			}
		})
	}
}

func TestPairList(t *testing.T) {
	for _, test := range []struct {
		data  map[string]string
		delim rune
		want  string
	}{
		{map[string]string{}, '→', ""},
		{map[string]string{" ": ""}, '→', ""},
		{map[string]string{"\t": ""}, '→', ""},
		{map[string]string{"a": ""}, '→', "A"},
		{map[string]string{"a": " "}, '→', "A"},
		{map[string]string{"a": "\t"}, '→', "A"},
		{map[string]string{"": "z"}, '→', ""},
		{map[string]string{" ": "z"}, '→', ""},
		{map[string]string{"\t": "z"}, '→', ""},
		{map[string]string{"a": "z"}, '→', "A→Z"},
		{map[string]string{"a": "z"}, ':', "A:Z"},
		{map[string]string{"a": "→z"}, '→', "A→→Z"},
		{map[string]string{"a": "", "b": ""}, '→', "A B"},
		{map[string]string{"a": " ", "b": "\t"}, '→', "A B"},
		{map[string]string{"a": "z", "b": "y"}, '→', "A→Z B→Y"},
		{map[string]string{"a": " z ", "b": "\ty\t"}, '→', "A→Z B→Y"},
		{map[string]string{"z": "a", "y": "b"}, '→', "Y→B Z→A"},
		{map[string]string{"a": "y z"}, '→', "A→YZ"},
		{map[string]string{"a": "z", "b": "y"}, ' ', "A Z B Y"},

		// Actual exit data
		{
			map[string]string{"N": "L1", "NE": "L3", "E": "L4"}, '→',
			"E→L4 NE→L3 N→L1",
		},

		//
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := PairList(test.data, test.delim)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

func BenchmarkPairList(b *testing.B) {
	for _, test := range []struct {
		name  string
		data  map[string]string
		delim rune
	}{
		{"ASCII delim", map[string]string{"a": "b"}, '→'},
		{"Unicode Delim", map[string]string{"a": "b"}, ':'},
		{"Exits x1", map[string]string{"N": "L1"}, '→'},
		{"Exits x2", map[string]string{"N": "L1", "NE": "L3"}, '→'},
		{"Exits x3", map[string]string{"N": "L1", "NE": "L3", "E": "L4"}, '→'},
		{
			"Door",
			map[string]string{"EXIT": "E", "RESET": "1m", "JITTER": "1m", "OPEN": ""},
			'→',
		},
		{"Action", map[string]string{"AFTER": "15s", "JITTER": "15s"}, '→'},
		{"Reset", map[string]string{"AFTER": "0s", "JITTER": "12m", "SPAWN": ""}, '→'},
	} {
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = PairList(test.data, test.delim)
			}
		})
	}
}

func TestStringList(t *testing.T) {
	for _, test := range []struct {
		data []string
		want string
	}{
		{[]string{}, ""},
		{[]string{" a", "b ", " c "}, "a\n: b\n: c"},
		{[]string{"c", "b", "a"}, "a\n: b\n: c"},

		// Actual OnAction data
		{
			[]string{
				"The frog croaks a bit.",
				"The little frog leaps high into the air.",
				"The frog hops around a bit.",
			},
			"The frog croaks a bit.\n" +
				": The frog hops around a bit.\n" +
				": The little frog leaps high into the air.",
		},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := StringList(test.data)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

func BenchmarkStringList(b *testing.B) {
	for _, test := range []struct {
		name string
		data []string
	}{
		{"OnAction x1", []string{"The frog croaks a bit."}},
		{"OnAction x2", []string{
			"The frog croaks a bit.",
			"The little frog leaps high into the air.",
		}},
		{"OnAction x3", []string{
			"The frog croaks a bit.",
			"The little frog leaps high into the air.",
			"The frog hops around a bit.",
		}},

		// Actually a KeyedStringList but it can be split using StringList for
		// benchmarking
		{"Veto x3", []string{
			"GET→The rock seems quite immovable.",
			"PUT→You can't put the rock anywhere.",
			"TAKE→You can't take the rock anywhere.",
		}},
	} {
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = StringList(test.data)
			}
		})
	}
}

func TestKeyedStringList(t *testing.T) {
	for _, test := range []struct {
		data  map[string]string
		delim rune
		want  string
	}{
		{map[string]string{}, '→', ""},
		{map[string]string{"": " "}, '→', ""},
		{map[string]string{" ": ""}, '→', ""},
		{map[string]string{" ": " "}, '→', ""},
		{map[string]string{"": "\t"}, '→', ""},
		{map[string]string{"\t": ""}, '→', ""},
		{map[string]string{"\t": "\t"}, '→', ""},
		{map[string]string{"a": ""}, '→', "A"},
		{map[string]string{"a": " "}, '→', "A"},
		{map[string]string{"a": "\t"}, '→', "A"},
		{map[string]string{"": "z"}, '→', ""},
		{map[string]string{" ": "z"}, '→', ""},
		{map[string]string{"\t": "z"}, '→', ""},
		{map[string]string{"a": "z"}, '→', "A→z"},
		{map[string]string{"a": "z", "b": "y"}, '→', "A→z\n: B→y"},
		{
			map[string]string{"a": "z", "b": "y", "c": "x"},
			'→', "A→z\n: B→y\n: C→x",
		},
		{
			map[string]string{"c": "x", "b": "y", "a": "z"},
			'→', "A→z\n: B→y\n: C→x",
		},

		// Real vetoes data
		{
			map[string]string{
				"GET": "The rock seems quite immovable.",
				"PUT": "You can't put the rock anywhere.",
			},
			'→',
			"GET→The rock seems quite immovable.\n" +
				": PUT→You can't put the rock anywhere.",
		},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := KeyedStringList(test.data, test.delim)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

func BenchmarkKeyedStringList(b *testing.B) {
	for _, test := range []struct {
		name  string
		data  map[string]string
		delim rune
	}{
		{"Veto x1",
			map[string]string{
				"GET": "The rock seems quite immovable.",
			}, '→',
		},
		{"Veto x2",
			map[string]string{
				"GET": "The rock seems quite immovable.",
				"PUT": "You can't put the rock anywhere.",
			}, '→',
		},
		{"Veto x3",
			map[string]string{
				"GET":  "The rock seems quite immovable.",
				"PUT":  "You can't put the rock anywhere.",
				"TAKE": "You can't take the rock anywhere.",
			}, '→',
		},
	} {
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = KeyedStringList(test.data, test.delim)
			}
		})
	}
}

func TestBytes(t *testing.T) {
	for _, test := range []struct {
		data string
		want string
	}{
		// Basic tests
		{"", ""},
		{" ", ""},
		{"\t", ""},
		{"\n", "\n"},
		{"\n\t\n", "\n\t\n"},
		{"\t\n\t", "\n"},
		{"Some text", "Some text"},

		// Leading white space
		{" Leading space", "Leading space"},
		{"\tLeading tab", "Leading tab"},
		{"\nLeading LF", "\nLeading LF"},
		{" \nLeading space+LF", "\nLeading space+LF"},
		{"\n Leading LF+space", "\n Leading LF+space"},
		{"\t\nLeading tab+LF", "\nLeading tab+LF"},
		{"\n\tLeading LF+tab", "\n\tLeading LF+tab"},
		{" \n Leading space+LF+space", "\n Leading space+LF+space"},
		{"\t\n\tLeading tab+LF+tab", "\n\tLeading tab+LF+tab"},

		// Trailing white space
		{"Trailing space ", "Trailing space"},
		{"Trailing tab\t", "Trailing tab"},
		{"Trailing LF\n", "Trailing LF\n"},
		{"Trailing LF+space\n ", "Trailing LF+space\n"},
		{"Trailing space+LF \n", "Trailing space+LF \n"},
		{"Trailing LF+tab\n\t", "Trailing LF+tab\n"},
		{"Trailing tab+LF\t\n", "Trailing tab+LF\t\n"},
		{"Trailing space+LF+space \n ", "Trailing space+LF+space \n"},
		{"Trailing tab+LF+tab\t\n\t", "Trailing tab+LF+tab\t\n"},

		// Leading and trailing white space (same both ends)
		{" Both space ", "Both space"},
		{"\tBoth tab\t", "Both tab"},
		{"\nBoth LF\n", "\nBoth LF\n"},
		{" \nBoth LF+space\n ", "\nBoth LF+space\n"},
		{"\t\nBoth LF+tab\n\t", "\nBoth LF+tab\n"},

		// Leading and trailing white space (RHS mirror of LHS)
		{
			"\n Leading LF+space, Trailing space+LF \n",
			"\n Leading LF+space, Trailing space+LF \n",
		},
		{
			"\n\tLeading LF+tab, Trailing tab+LF\t\n",
			"\n\tLeading LF+tab, Trailing tab+LF\t\n",
		},
		{
			" \n Leading space+LF+space, Trailing space+LF+space \n ",
			"\n Leading space+LF+space, Trailing space+LF+space \n",
		},
		{
			"\t\n\tLeading tab+LF+tab, Trailing tab+LF+tab\t\n\t",
			"\n\tLeading tab+LF+tab, Trailing tab+LF+tab\t\n",
		},

		// Real data, description of tavern fireplace
		{
			"You are in the corner of the common room in the dragon's breath tavern. A fire\nburns merrily in an ornate fireplace, giving comfort to weary travellers. The\nfire causes shadows to flicker and dance around the room, changing darkness to\nlight and back again. To the south the common room continues and east the common\nroom leads to the tavern entrance.",
			"You are in the corner of the common room in the dragon's breath tavern. A fire\nburns merrily in an ornate fireplace, giving comfort to weary travellers. The\nfire causes shadows to flicker and dance around the room, changing darkness to\nlight and back again. To the south the common room continues and east the common\nroom leads to the tavern entrance.",
		},
	} {
		t.Run(fmt.Sprintf("%s", test.want), func(t *testing.T) {
			have := Bytes([]byte(test.data))

			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

func TestBytesSideEfects(t *testing.T) {

	// Setup test with access to backing array for checking
	const sample = " Some Text "
	data := [len(sample)]byte{}
	copy(data[:], sample)
	test := data[:]

	have := Bytes(test)

	// Make sure passed test data isn't accidentally modified
	if !bytes.Equal(test, []byte(sample)) {
		t.Errorf("passed parameter modified\nhave: %+q\nwant: %+q",
			test, sample,
		)
	}

	// Overwrite the returned result...
	have = have[:cap(have)]
	for x := range have {
		have[x] = 0x00
	}

	// ...If the passed test data is not equal to our initial sample then we
	// havn't had a copy returned as overwriting the result modified the sample
	if !bytes.Equal(data[:], []byte(sample)) {
		t.Errorf("copy not returned\nhave: %+q\nwant: %+q",
			data[:], sample,
		)
	}
}

func BenchmarkBytes(b *testing.B) {
	for _, test := range []struct {
		name string
		data string
	}{
		{"Description", "You are in the corner of the common room in the dragon's breath tavern. A fire\nburns merrily in an ornate fireplace, giving comfort to weary travellers. The\nfire causes shadows to flicker and dance around the room, changing darkness to\nlight and back again. To the south the common room continues and east the common\nroom leads to the tavern entrance."},
	} {
		data := []byte(test.data)
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Bytes(data)
			}
		})
	}
}

func TestDuration(t *testing.T) {
	for _, test := range []struct {
		duration string
		want     string
	}{
		{"0", "0s"},
		{"100ms", "0s"},
		{"0.1s", "0s"},
		{"0.5s", "1s"},
		{"0.9s", "1s"},
		{"1s", "1s"},
		{"60s", "1m"},
		{"1m", "1m"},
		{"1m0s", "1m"},
		{"1h", "1h"},
		{"1h0s", "1h"},
		{"1h0m", "1h"},
		{"1h0m0s", "1h"},
		{"1h0m1s", "1h1s"},
		{"0h1m0s", "1m"},
		{"1h1m1s", "1h1m1s"},
		{"1.5h", "1h30m"},
		{"0h1m0s", "1m"},
	} {
		t.Run(fmt.Sprintf("%s", test.duration), func(t *testing.T) {
			d, err := time.ParseDuration(test.duration)
			if err != nil {
				t.Errorf("invalid duration: %s", test.duration)
			}
			have := Duration(d)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

func TestDateTime(t *testing.T) {

	UTC := time.FixedZone("UTC", 0)
	refdt := time.Date(2018, time.September, 20, 20, 24, 33, 0, UTC)
	want := []byte("Thu, 20 Sep 2018 20:24:33 +0000")

	for _, offset := range []int{
		0, 5, -5,
	} {
		t.Run(fmt.Sprintf("%d", offset), func(t *testing.T) {

			// Get reference date/time in test timezone
			zoneName := fmt.Sprintf("UTC%+d", offset)
			zone := time.FixedZone(zoneName, offset*60*60)
			dt := refdt.In(zone)

			have := DateTime(dt)

			if !bytes.Equal(have, want) {
				t.Errorf("\nhave %+q\nwant %+q", have, want)
			}
		})
	}
}

func TestBoolean(t *testing.T) {
	for _, test := range []struct {
		data bool
		want string
	}{
		{true, "TRUE"},
		{false, "FALSE"},
	} {
		t.Run(fmt.Sprintf("%s", test.want), func(t *testing.T) {
			have := Boolean(test.data)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

func TestInteger(t *testing.T) {
	for _, test := range []struct {
		data int
		want string
	}{
		{0, "0"},
		{-0, "0"},
		{123456789, "123456789"},
		{-123456789, "-123456789"},
	} {
		t.Run(fmt.Sprintf("%d", test.data), func(t *testing.T) {
			have := Integer(test.data)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}
