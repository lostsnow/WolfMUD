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
		{"spaces-1", " some text "},
		{"spaces-2", "  some text  "},
		{"tabs-1", "\tsome text\t"},
		{"tabs-2", "\t\tsome text\t\t"},
		{"mixed", "\t some text \t"},
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
		{"split", "split keyword"},
	} {
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = Keyword(test.keyword)
			}
		})
	}
}

func Test_KeywordList(t *testing.T) {
	for _, test := range []struct {
		data []string
		want string
	}{
		{[]string{}, ""},
		{[]string{""}, ""},
		{[]string{" "}, ""},
		{[]string{"", ""}, ""},
		{[]string{" ", " "}, ""},
		{[]string{"a", "keyword", "test"}, "A KEYWORD TEST"},
		{[]string{" a", "keyword ", " test "}, "A KEYWORD TEST"},
		{[]string{"  a", "keyword  ", "  test  "}, "A KEYWORD TEST"},
		{[]string{"\ta", "keyword\t", "\ttest\t"}, "A KEYWORD TEST"},
		{[]string{"spaced keyword"}, "SPACEDKEYWORD"},
		{[]string{"spaced  keyword"}, "SPACEDKEYWORD"},
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
		{"Real", []string{"L2N1", "L2N2", "L2N3", "M4"}},
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
		{map[string]string{"a": ""}, '→', "A"},
		{map[string]string{"a": "z"}, '→', "A→Z"},
		{map[string]string{"a": "z"}, ':', "A:Z"},
		{map[string]string{"a": "", "b": ""}, '→', "A B"},
		{map[string]string{"a": "z", "b": "y"}, '→', "A→Z B→Y"},
		{map[string]string{"z": "a", "y": "b"}, '→', "Y→B Z→A"},
		{map[string]string{"a": "y z"}, '→', "A→YZ"},
		{map[string]string{"a": "z", "b": "y"}, ' ', "A Z B Y"},

		// Actual data
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
		{"2x1", map[string]string{"a": "z", "b": "y"}, '→'},
		{"3x1", map[string]string{"a": "z", "b": "y", "c": "x"}, '→'},
		{"3x3", map[string]string{"aaa": "z", "bbb": "y", "ccc": "x"}, '→'},
		{"3x6", map[string]string{"aaaaaa": "z", "bbbbbb": "y", "cccccc": "x"}, '→'},
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
		{
			[]string{"the quick brown", "fox jumps over", "the lazy dog."},
			"fox jumps over\n: the lazy dog.\n: the quick brown",
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
		{"3x1", []string{"a", "b", "c"}},
		{"3x3", []string{"aaa", "bbb", "ccc"}},
		{"3x6", []string{"aaaaaa", "bbbbbb", "cccccc"}},
		// Same line + folding
		{"4x10", []string{"aaaaaaaaaa", "bbbbbbbbbb", "cccccccccc", "dddddddddd"}},
		// "\n: " separator for each item
		{"5x10", []string{
			"aaaaaaaaaa", "bbbbbbbbbb", "cccccccccc", "dddddddddd", "eeeeeeeeee",
		}},
	} {
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = StringList(test.data)
			}
		})
	}
}

func TestKeyedString(t *testing.T) {
	for _, test := range []struct {
		key   string
		value string
		delim rune
		want  string
	}{
		{"", "", '→', ""},
		{"", "Some text", '→', "→Some text"}, // correct?
		{"a", "", '→', "A"},
		{"b", "Some text", '→', "B→Some text"},
		{" c", "Some text", '→', "C→Some text"},
		{"d ", "Some text", '→', "D→Some text"},
		{" e ", "Some text", '→', "E→Some text"},
		{"f", " Some text", '→', "F→Some text"},
		{"g", "Some text ", '→', "G→Some text"},
		{"h ", " Some text ", '→', "H→Some text"},
		{"i i i", " Some text ", '→', "III→Some text"},
	} {
		t.Run(fmt.Sprintf("%s", test.key), func(t *testing.T) {
			have := KeyedString(test.key, test.value, test.delim)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
			}
		})
	}
}

func BenchmarkKeyedString(b *testing.B) {
	for _, test := range []struct {
		name  string
		key   string
		value string
		delim rune
	}{
		{"no key", "", "Some text", '→'},
		{"no value", "a", "", '→'},
		{"no trim", "a", "Some text", '→'},
		{"trim key", " a ", "Some text", '→'},
		{"trim value", "a", " Some text ", '→'},
		{"trim both", " a ", " Some text ", '→'},
	} {
		b.Run(fmt.Sprintf(test.name), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = KeyedString(test.key, test.value, test.delim)
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
		{map[string]string{"a": ""}, '→', "A"},
		{map[string]string{"": "z"}, '→', "→z"}, // correct?
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
		{"x1", map[string]string{"a": "z"}, '→'},
		{"x2", map[string]string{"a": "z", "b": "y"}, '→'},
		{"ordered", map[string]string{"a": "z", "b": "y", "c": "x"}, '→'},
		{"unordered", map[string]string{"c": "x", "b": "y", "a": "z"}, '→'},
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
		data []byte
	}{
		{[]byte("Some text")},
	} {
		t.Run(fmt.Sprintf("%s", test.data), func(t *testing.T) {
			want := test.data
			have := Bytes(test.data)

			// Take address of last element [cap(x)-1] from the maximum sized slice
			// x[0:cap(x)] and if they are the same then slices overlap
			haveEnd := &have[0:cap(have)][cap(have)-1]
			wantEnd := &want[0:cap(want)][cap(want)-1]
			if haveEnd == wantEnd {
				t.Errorf("have and want overlap: %+q", have)
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
	for _, test := range []struct {
		yyyy              int
		mm                time.Month
		dd, h, m, s, nsec int
		loc               string
		want              string
	}{
		{1971, time.January, 1, 11, 59, 59, 0, "UTC", "Fri, 01 Jan 1971 11:59:59 UTC"},
		{1971, time.January, 1, 11, 59, 59, 0, "GMT", "Fri, 01 Jan 1971 11:59:59 UTC"},
		{1971, time.August, 1, 11, 59, 59, 0, "Europe/London", "Sun, 01 Aug 1971 10:59:59 UTC"},
	} {
		loc, _ := time.LoadLocation(test.loc)
		t.Run(fmt.Sprintf("%d", test.yyyy), func(t *testing.T) {
			d := time.Date(test.yyyy, test.mm, test.dd, test.h, test.m, test.s, test.nsec, loc)
			have := DateTime(d)
			if !bytes.Equal(have, []byte(test.want)) {
				t.Errorf("\nhave %+q\nwant %+q", have, test.want)
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
