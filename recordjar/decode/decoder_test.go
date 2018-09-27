// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package decode_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	. "code.wolfmud.org/WolfMUD.git/recordjar/decode"
)

func TestString(t *testing.T) {
	for _, test := range []struct {
		data string
		want string
	}{
		{"", ""},
		{"all lowercase", "all lowercase"},
		{"ALL UPPERCASE", "ALL UPPERCASE"},
		{"All Titlecase", "All Titlecase"},
		{"AlL MiXeDcAsE", "AlL MiXeDcAsE"},
		{" Leading Space", "Leading Space"},
		{"  Leading Space", "Leading Space"},
		{"Trailing Space ", "Trailing Space"},
		{"Trailing Space  ", "Trailing Space"},
		{" Both Space ", "Both Space"},
		{"  Both Space  ", "Both Space"},
		{"\tLeading Tab", "Leading Tab"},
		{"\t\tLeading Tab", "Leading Tab"},
		{"Trailing Tab\t", "Trailing Tab"},
		{"Trailing Tab\t\t", "Trailing Tab"},
		{"\tBoth Tab\t", "Both Tab"},
		{"\t\tBoth Tab\t\t", "Both Tab"},
		{"\t Leading Tab", "Leading Tab"},
		{"Trailing Tab\t ", "Trailing Tab"},
		{"\t Both Tab\t ", "Both Tab"},
		{" \tLeading Tab", "Leading Tab"},
		{"Trailing Tab \t", "Trailing Tab"},
		{" \tBoth Tab \t", "Both Tab"},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := String([]byte(test.data))
			if have != test.want {
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
		{"spaces-1", " some text "},
		{"spaces-2", "  some text  "},
		{"tabs-1", "\tsome text\t"},
		{"tabs-2", "\t\tsome text\t\t"},
		{"mixed", "\t some text \t"},
	} {
		data := []byte(test.keyword)
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = String(data)
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
		{"  ", ""},
		{"\t", ""},
		{"\t\t", ""},
		{"keyword", "KEYWORD"},
		{" keyword", "KEYWORD"},
		{"  keyword", "KEYWORD"},
		{"keyword ", "KEYWORD"},
		{"keyword  ", "KEYWORD"},
		{" keyword ", "KEYWORD"},
		{"  keyword  ", "KEYWORD"},
		{"\tkeyword", "KEYWORD"},
		{"keyword\t", "KEYWORD"},
		{"\tkeyword\t", "KEYWORD"},
		{"keyword\n", "KEYWORD"},
		{"spaced  keyword", "SPACEDKEYWORD"},
		{"spaced   keyword", "SPACEDKEYWORD"},
		{"spaced\tkeyword", "SPACEDKEYWORD"},
		{"spaced\t\tkeyword", "SPACEDKEYWORD"},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := Keyword([]byte(test.data))
			if have != test.want {
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
		{"split", "split keyword"},
	} {
		data := []byte(test.keyword)
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Keyword(data)
			}
		})
	}
}

func TestKeywordList(t *testing.T) {
	for _, test := range []struct {
		data string
		want []string
	}{
		{"", []string{}},
		{"", []string{}},
		{" ", []string{}},
		{"  ", []string{}},
		{"a keyword test", []string{"A", "KEYWORD", "TEST"}},
		{" a keyword test ", []string{"A", "KEYWORD", "TEST"}},
		{"  a keyword  test  ", []string{"A", "KEYWORD", "TEST"}},
		{"\ta keyword\t \ttest\t", []string{"A", "KEYWORD", "TEST"}},
		{"z y x", []string{"X", "Y", "Z"}},
		{"ABC ABC XYZ XYZ", []string{"ABC", "XYZ"}},
		{"ABC abc XYZ xyz", []string{"ABC", "XYZ"}},
		{"abc ABC xyz XYZ", []string{"ABC", "XYZ"}},
		{"ABC XYZ ABC XYZ", []string{"ABC", "XYZ"}},
		{"ABC\nABC\nXYZ\nXYZ", []string{"ABC", "XYZ"}},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := KeywordList([]byte(test.data))
			if len(have) != len(test.want) {
				t.Errorf("unequal slices\nhave %+q\nwant %+q", have, test.want)
				return
			}
			for x := range have {
				if have[x] != test.want[x] {
					t.Errorf("value missmatch\nhave %+q\nwant %+q", have, test.want)
					return
				}
			}
		})
	}
}

func BenchmarkKeywordList(b *testing.B) {
	for _, test := range []struct {
		name     string
		keywords string
	}{
		{"1x1", "a"},
		{"3x1", "c b a"},
		{"3x3", "ccc bbb aaa"},
		{"3x3Dup1", "ABC ABC XYZ"},
		{"3x3Dup2", "ABC XYZ XYZ"},
		{"3x3Dup3", "ABC ABC ABC"},
		{"6x1", "f e d c b a"},
		{"6x3", "fff eee ddd ccc bbb aaa"},
		{"6x10",
			"ffffffffff eeeeeeeeee dddddddddd cccccccccc bbbbbbbbbb aaaaaaaaaa",
		},
		{"Real", "L2N1 L2N2 L2N3 M4"},
	} {
		data := []byte(test.keywords)
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = KeywordList(data)
			}
		})
	}
}

func TestPairList(t *testing.T) {
	for _, test := range []struct {
		data string
		want map[string]string
	}{
		{"", map[string]string{}},
		{"a", map[string]string{"A": ""}},
		{"a→", map[string]string{"A": ""}},
		{"→z", map[string]string{}},
		{"→", map[string]string{}},
		{"→ →", map[string]string{}},
		{"a→z", map[string]string{"A": "Z"}},
		{"a→→z", map[string]string{"A": "→Z"}},
		{"a:z", map[string]string{"A": "Z"}},
		{"a b→", map[string]string{"A": "", "B": ""}},
		{"a→ a→", map[string]string{"A": ""}},
		{"a→z b→y", map[string]string{"A": "Z", "B": "Y"}},
		{"y→b z→a", map[string]string{"Z": "A", "Y": "B"}},
		{"a→zy", map[string]string{"A": "ZY"}},

		// Should only get first occurance of duplicate keyword
		{"a→z a→y", map[string]string{"A": "Z"}},

		// Actual data
		{
			"E→L4 NE→L3 N→L1",
			map[string]string{"N": "L1", "NE": "L3", "E": "L4"},
		},

		//
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := PairList([]byte(test.data))
			if len(have) != len(test.want) {
				t.Errorf("unequal maps\nhave %+q\nwant %+q", have, test.want)
				return
			}
			for x := range have {
				if _, ok := test.want[x]; !ok {
					t.Errorf("extra value\nhave %+q", have)
					continue
				}
			}
			for x := range test.want {
				if _, ok := have[x]; !ok {
					t.Errorf("missing value\nhave %+q", test.want)
					continue
				}
				if have[x] != test.want[x] {
					t.Errorf("\nhave %+q\nwant %+q", have[x], test.want[x])
				}
			}
		})
	}
}

func BenchmarkPairList(b *testing.B) {
	for _, test := range []struct {
		name  string
		pairs string
	}{
		{"Exits x1", "N→L14"},
		{"Exits x2", "N→L14 E→L6"},
		{"Exits x3", "N→L14 E→L6 S→L7"},
		{"Exits x4", "N→L14 E→L6 S→L7 W→L3"},
		{"Door", "EXIT→E RESET→1m JITTER→1m OPEN"},
		{"Action", "AFTER→15s JITTER→15s"},
		{"Reset", "AFTER→0s JITTER→2m SPAWN"},
	} {
		data := []byte(test.pairs)
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = PairList(data)
			}
		})
	}
}

func TestStringList(t *testing.T) {
	for _, test := range []struct {
		data string
		want []string
	}{
		{"", []string{}},
		{" ", []string{}},
		{"a", []string{"a"}},
		{"a:", []string{"a"}},
		{":a", []string{"a"}},
		{"a:b", []string{"a", "b"}},
		{"a:b:", []string{"a", "b"}},
		{":a:b", []string{"a", "b"}},
		{":a:b:", []string{"a", "b"}},
		{"a : b", []string{"a", "b"}},
		{" a : b ", []string{"a", "b"}},
		{"a b : c d", []string{"a b", "c d"}},
		{": a\n: b", []string{"a", "b"}},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := StringList([]byte(test.data))
			if len(have) != len(test.want) {
				t.Errorf("unequal slices\nhave %+q\nwant %+q", have, test.want)
				return
			}
			for x := range have {
				if have[x] != test.want[x] {
					t.Errorf("value missmatch\nhave %+q\nwant %+q", have, test.want)
					return
				}
			}
		})
	}
}

func BenchmarkStringList(b *testing.B) {
	for _, test := range []struct {
		name    string
		strings string
	}{
		{"OnAction x1", "The rabbit hops around a bit."},
		{"OnAction x2", "The rabbit hops around a bit. : You see the rabbit twitch its little nose, Ahh..."},
		{"OnAction x3", "The rabbit hops around a bit. : You see the rabbit twitch its little nose, Ahh... : The rabbit makes a soft squeaking and chattering noise."},

		// Actually a KeyedStringList which also uses StringList to split the list
		{"Veto x3", "GET→The rock seems quite immovable. : PUT→You can't put the rock anywhere. : TAKE→You can't take the rock anywhere."},
	} {
		data := []byte(test.strings)
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = StringList(data)
			}
		})
	}
}

func TestKeyedStringList(t *testing.T) {
	for _, test := range []struct {
		data string
		want map[string]string
	}{
		{"", map[string]string{}},
		{":", map[string]string{}},
		{" : ", map[string]string{}},
		{"::", map[string]string{}},
		{" : : ", map[string]string{}},
		{"a", map[string]string{"A": ""}},
		{"a→", map[string]string{"A": ""}},
		{"a→z", map[string]string{"A": "z"}},
		{" a→z ", map[string]string{"A": "z"}},
		{" a → z ", map[string]string{"A": "z"}},
		{"a→z y", map[string]string{"A": "z y"}},
		{"a b→z y", map[string]string{"AB": "z y"}},
		{"a→z:b→y", map[string]string{"A": "z", "B": "y"}},
		{"a b→z:c d→y", map[string]string{"AB": "z", "CD": "y"}},
		{"a→z y:b→x w", map[string]string{"A": "z y", "B": "x w"}},
		{"a b→z y:c d→x w", map[string]string{"AB": "z y", "CD": "x w"}},
		{"a→z : b→y", map[string]string{"A": "z", "B": "y"}},
		{"a b→z : c d→y", map[string]string{"AB": "z", "CD": "y"}},
		{"a→z y : b→x w ", map[string]string{"A": "z y", "B": "x w"}},
		{"a b→z y : c d→x w ", map[string]string{"AB": "z y", "CD": "x w"}},
		{"a → z y : b → x w", map[string]string{"A": "z y", "B": "x w"}},
		{"a b → z y : c d → x w", map[string]string{"AB": "z y", "CD": "x w"}},
		{":a→z y \n:b→x w", map[string]string{"A": "z y", "B": "x w"}},
		{":a→z y \n:b→x w", map[string]string{"A": "z y", "B": "x w"}},
		{": a → z y \n: b → x w", map[string]string{"A": "z y", "B": "x w"}},
		{": a b → z y \n: c d → x w", map[string]string{"AB": "z y", "CD": "x w"}},

		// Should only get first occurance of duplicate keyword
		{"a→z:a→y", map[string]string{"A": "z"}},

		// Real vetoes data:
		{
			"GET→The rock seems quite immovable. : PUT→You can't put the rock anywhere.",
			map[string]string{
				"GET": "The rock seems quite immovable.",
				"PUT": "You can't put the rock anywhere.",
			}},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := KeyedStringList([]byte(test.data))
			if len(have) != len(test.want) {
				t.Errorf("unequal maps\nhave %+q\nwant %+q", have, test.want)
				return
			}
			for x := range have {
				if _, ok := test.want[x]; !ok {
					t.Errorf("extra value\nhave %+q", have)
					continue
				}
			}
			for x := range test.want {
				if _, ok := have[x]; !ok {
					t.Errorf("missing value\nhave %+q", test.want)
					continue
				}
				if have[x] != test.want[x] {
					t.Errorf("\nhave %+q\nwant %+q", have[x], test.want[x])
				}
			}
		})
	}
}

func BenchmarkKeyedStringList(b *testing.B) {
	for _, test := range []struct {
		name string
		data string
	}{
		{"Veto x1", "GET→The rock seems quite immovable."},
		{"Veto x2", "GET→The rock seems quite immovable. : PUT→You can't put the rock anywhere."},
		{"Veto x3", "GET→The rock seems quite immovable. : PUT→You can't put the rock anywhere. : TAKE→You can't take the rock anywhere."},
	} {
		data := []byte(test.data)
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = KeyedStringList(data)
			}
		})
	}
}

func TestBytes(t *testing.T) {
	for _, test := range []struct {
		data []byte
		want []byte
	}{
		{[]byte(""), []byte("")},
		{[]byte(" "), []byte("")},
		{[]byte("a"), []byte("a")},
		{[]byte(" a "), []byte("a")},
		{[]byte("\na\n"), []byte("\na\n")},
		{[]byte(" \n a \n "), []byte("\n a \n")},

		// Real data, description of tavern fireplace
		{[]byte("You are in the corner of the common room in the dragon's breath tavern. A fire\nburns merrily in an ornate fireplace, giving comfort to weary travellers. The\nfire causes shadows to flicker and dance around the room, changing darkness to\nlight and back again. To the south the common room continues and east the common\nroom leads to the tavern entrance."), []byte("You are in the corner of the common room in the dragon's breath tavern. A fire\nburns merrily in an ornate fireplace, giving comfort to weary travellers. The\nfire causes shadows to flicker and dance around the room, changing darkness to\nlight and back again. To the south the common room continues and east the common\nroom leads to the tavern entrance.")},

		//
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			orig := make([]byte, len(test.data))
			copy(orig, test.data)

			have := Bytes(test.data)

			// Make sure test data isn't accidentally modified
			if !bytes.Equal(orig, test.data) {
				t.Errorf("passed parameter modified\nhave: %+q\nwant: %+q",
					orig, test.data,
				)
			}

			if !bytes.Equal(have, test.want) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}

			// Overwrite what we have, if the test data changes then we havn't had a
			// copy returned
			have = have[:cap(have)]
			for x := range have {
				have[x] = 0x00
			}
			if !bytes.Equal(orig, test.data) {
				t.Errorf("copy not returned\nhave: %+q\nwant: %+q", test.data, orig)
			}

		})
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
		data string
		want time.Duration
	}{
		{"", 0},
		{"unparseable", 0},
		{"1s", 1 * time.Second},
		{"1S", 1 * time.Second},
		{"1m", 1 * time.Minute},
		{"1M", 1 * time.Minute},
		{"1h", 1 * time.Hour},
		{"1H", 1 * time.Hour},
		{"1h2m3s", 1*time.Hour + 2*time.Minute + 3*time.Second},
		{" 1h2m3s ", 1*time.Hour + 2*time.Minute + 3*time.Second},
		{"1h 2m 3s", 1*time.Hour + 2*time.Minute + 3*time.Second},
		{"1h30m", 90 * time.Minute},
		{"1h30s", 1*time.Hour + 30*time.Second},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := Duration([]byte(test.data))
			if have != test.want {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
				return
			}
		})
	}
}

func BenchmarkDuration(b *testing.B) {
	for _, test := range []struct {
		name string
		data string
	}{
		{"second", "1s"},
		{"trim s", " 1s "},
		{"minute", "1m"},
		{"min+sec", "1m1s"},
		{"trim ms", " 1m1s "},
		{"WS ms", " 1m 1s"},
		{"WS+trim ms", " 1m 1s "},
		{"hour", "1h"},
		{"hour+minute", "1h1m"},
		{"hour+second", "1h1s"},
		{"hour+min+sec", "1h1s"},
		{"trim hms", " 1h1m1s "},
		{"WS hms", "1h 1m 1s"},
		{"WS+trim hms", " 1h 1m 1s "},
	} {
		data := []byte(test.data)
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Duration(data)
			}
		})
	}
}

func TestDateTime(t *testing.T) {

	UTC := time.FixedZone("UTC", 0)
	want := time.Date(2018, time.September, 20, 20, 24, 33, 0, UTC)

	for _, test := range []string{
		// Legacy pre WolfMUD v0.0.11 RFC1123 format
		"Thu, 20 Sep 2018 20:24:33 UTC",
		"Thu, 20 Sep 2018 21:24:33 BST",
		// Newer RFC1123Z format
		"Thu, 20 Sep 2018 20:24:33 +0000",
		"Thu, 20 Sep 2018 21:24:33 +0100",
		"Thu, 21 Sep 2018 01:24:33 +0500",
		"Thu, 20 Sep 2018 15:24:33 -0500",
	} {
		t.Run(fmt.Sprintf("%s", test), func(t *testing.T) {
			have := DateTime([]byte(test))
			if !have.Equal(want) {
				t.Errorf("\nhave %12d %+q\nwant %12d %+q",
					have.Unix(), have, want.Unix(), want,
				)
				return
			}
		})
	}
}

func TestBoolean(t *testing.T) {
	for _, test := range []struct {
		data string
		want bool
	}{
		{"", true},
		{"0", false},
		{" 0 ", false},
		{"f", false},
		{"F", false},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{" FALSE ", false},
		{"1", true},
		{" 1 ", true},
		{"t", true},
		{"T", true},
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{" TRUE ", true},
		{"invalid", false},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := Boolean([]byte(test.data))
			if have != test.want {
				t.Errorf("\nhave %t\nwant %t", have, test.want)
				return
			}
		})
	}
}

func TestInteger(t *testing.T) {
	for _, test := range []struct {
		data string
		want int
	}{
		{"", 0},
		{"-0", 0},
		{"+0", 0},
		{"1", 1},
		{"-1", -1},
		{"+1", +1},
		{"-2147483648", -2147483648}, // Minimum
		{"2147483647", 2147483647},   // Maximum
		{"invalid", 0},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			have := Integer([]byte(test.data))
			if have != test.want {
				t.Errorf("\nhave %d\nwant %d", have, test.want)
				return
			}
		})
	}
}
