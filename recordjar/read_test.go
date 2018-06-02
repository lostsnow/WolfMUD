// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	. "code.wolfmud.org/WolfMUD.git/recordjar"
)

// compare is a helper to compare two Jars j1 and j2. Parameter n can be used
// to identify which jar in a number of jars is being compared.
func compare(t *testing.T, id string, j1, j2 Jar) {

	const (
		extra   = "has extra"
		missing = "is missing"
	)

	t.Helper()
	f := func(reason string) {
		t.Helper()
		for x, r := range j1 {
			if x > len(j2)-1 {
				t.Errorf("jar %q, %s record %d", id, reason, x)
				continue
			}
			for field, value := range r {
				if _, ok := j2[x][field]; !ok {
					t.Errorf("jar %q, record %d - output %s field %q", id, x, reason, field)
					continue
				}
				if reason == extra && !bytes.Equal(value, j2[x][field]) {
					t.Errorf("jar %q, record %d, field: %q\nhave: %q\nwant: %q", id, x, field, value, j2[x][field])
				}
			}
		}
	}
	// Compare j1 with j2, then j2 with j1. First compare will report extra records
	// and fields, second compare will report missing records and fields.
	f(missing)
	j1, j2 = j2, j1
	f(extra)
}

// Test simple data from strings being parsed into Jars.
func TestRead_strings(t *testing.T) {
	for x, test := range []struct {
		data string
		want Jar
	}{
		// Empty jars
		{"", Jar{}},
		{"%%", Jar{}},
		{"  %%", Jar{}},
		{"%%  ", Jar{}},
		{"  %%  ", Jar{}},
		{"%%\n%%", Jar{}},
		{"// Comment\n%%", Jar{}},

		// Single field
		{"F1: d1", Jar{Record{"F1": []byte("d1")}}},
		{"F1:d1", Jar{Record{"F1": []byte("d1")}}},
		{"// Comment\nF1: d1", Jar{Record{"F1": []byte("d1")}}},
		{"// Comment\nF1:d1", Jar{Record{"F1": []byte("d1")}}},
		{"// Comment\nF1: d1\n%%", Jar{Record{"F1": []byte("d1")}}},
		{"// Comment\nF1:d1\n%%", Jar{Record{"F1": []byte("d1")}}},

		// Lowercased single field
		{"f1: d1", Jar{Record{"F1": []byte("d1")}}},
		{"f1:d1", Jar{Record{"F1": []byte("d1")}}},

		// Field prefixed with whitespace
		{"// Comment\n  F1: d1\n%%\n", Jar{Record{"F1": []byte("d1")}}},

		// Field with trailing line feed
		{"// Comment\nF1: d1\n%%\n", Jar{Record{"F1": []byte("d1")}}},

		// Field split over multiple lines
		{"F1: d1a\n    d1b", Jar{Record{"F1": []byte("d1a d1b")}}},

		// Duplicate field names
		{"F1: d1a\nF1: d1b", Jar{Record{"F1": []byte("d1a d1b")}}},

		// Whitespace around separator
		{"f1: d1\n  %%", Jar{Record{"F1": []byte("d1")}}},
		{"f1: d1\n%%  ", Jar{Record{"F1": []byte("d1")}}},
		{"f1: d1\n  %%  ", Jar{Record{"F1": []byte("d1")}}},
		{"f1: d1\n\t%%", Jar{Record{"F1": []byte("d1")}}},
		{"f1: d1\n%%\t", Jar{Record{"F1": []byte("d1")}}},
		{"f1: d1\n\t%%\t", Jar{Record{"F1": []byte("d1")}}},

		// Indented field
		{"  f1: d1", Jar{Record{"F1": []byte("d1")}}},
		{"\tf1: d1", Jar{Record{"F1": []byte("d1")}}},
		{"\t  f1: d1", Jar{Record{"F1": []byte("d1")}}},
		{"  \tf1: d1", Jar{Record{"F1": []byte("d1")}}},

		// Fields with \r\n line endings
		{"F1: d1a\r\nd1b\r\nF2: d2\r\n\r\nSome text three.\r\n\r\n\r\n%%",
			Jar{
				Record{
					"F1":       []byte("d1a d1b"),
					"F2":       []byte("d2"),
					"FREETEXT": []byte("Some text three.\n\n"),
				},
			},
		},

		// Multiple records
		{"  F1:D1\n%%\n  F2:D2\n%%\n", Jar{
			Record{"F1": []byte("D1")},
			Record{"F2": []byte("D2")},
		}},

		// Multiple records and freetext + ending separator
		{"F1:D1\n\nThe quick brown fox\n%%\nF2:D2\n\njumps over the lazy dog.\n%%\n",
			Jar{
				Record{
					"F1":       []byte("D1"),
					"FREETEXT": []byte("The quick brown fox"),
				},
				Record{
					"F2":       []byte("D2"),
					"FREETEXT": []byte("jumps over the lazy dog."),
				},
			},
		},

		// Multiple records and freetext, with NO ending separator
		{"F1:D1\n\nThe quick brown fox\n%%\nF2:D2\n\njumps over the lazy dog.\n", Jar{
			Record{
				"F1":       []byte("D1"),
				"FREETEXT": []byte("The quick brown fox"),
			},
			Record{
				"F2":       []byte("D2"),
				"FREETEXT": []byte("jumps over the lazy dog."),
			},
		}},

		// Multiple records and freetext, with NO ending separator or new line
		{"F1:D1\n\nThe quick brown fox\n%%\nF2:D2\n\njumps over the lazy dog.", Jar{
			Record{
				"F1":       []byte("D1"),
				"FREETEXT": []byte("The quick brown fox"),
			},
			Record{
				"F2":       []byte("D2"),
				"FREETEXT": []byte("jumps over the lazy dog."),
			},
		}},
	} {
		t.Run(strconv.Itoa(x), func(t *testing.T) {
			have := Read(bytes.NewBufferString(test.data), "freetext")
			compare(t, strconv.Itoa(x), have, test.want)
		})
	}
}

var greeting = Record{
	"FREETEXT": []byte(`
WolfMUD Copyright 1984-2016 Andrew 'Diddymus' Rolfe

    World
    Of
    Living
    Fantasy

Welcome to WolfMUD!
`),
}

var location = Record{
	"REF":       []byte("L1"),
	"START":     []byte(""),
	"NAME":      []byte("Fireplace"),
	"ALIASES":   []byte("TAVERN FIREPLACE"),
	"EXITS":     []byte("E→L3 SE→L4 S→L2"),
	"INVENTORY": []byte("L1N1"),
	"FREETEXT":  []byte("You are in the corner of the common room in the dragon's breath tavern. A fire burns merrily in an ornate fireplace, giving comfort to weary travellers. The fire causes shadows to flicker and dance around the room, changing darkness to light and back again. To the south the common room continues and east the common room leads to the tavern entrance."),
}

// Test larger data from files being parsed into Jars.
func TestRead_files(t *testing.T) {
	for _, test := range []struct {
		filename string
		want     Jar
	}{
		// Plain server greeting with blanks lines and indents, no comment or %%
		{"greeting.wrj", Jar{greeting}},
		// Server greeting with blanks lines and indents, with comment
		{"greeting-comment.wrj", Jar{greeting}},
		// Server greeting with blanks lines and indents, no comment
		{"greeting-nocomment.wrj", Jar{greeting}},
		// Sample location
		{"location.wrj", Jar{location}},
		// Sample location with record separator and DOS line endings
		{"location-nosep.wrj", Jar{location}},
		// Sample location without record separator
		{"location-dos.wrj", Jar{location}},
		// Sample location with space indents before comments and separator
		{"location-indent-space.wrj", Jar{location}},
		// Sample location with tab indents before comments and separator
		{"location-indent-tab.wrj", Jar{location}},
	} {
		t.Run(test.filename, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", test.filename))
			if err != nil {
				t.Fatalf("%s", err)
			}

			have := Read(f, "freetext")
			compare(t, test.filename, have, test.want)

			f.Close()
		})
	}
}

// Test freetext data from files being parsed into Jars. This is easier with
// files than with string literals.
func TestRead_freetext(t *testing.T) {
	for x, test := range []struct {
		filename string
		want     string
	}{
		{"ft-plain.wrj", "The quick brown fox jumps over the lazy dog."},
		{"ft-embed-blank.wrj", "The quick brown fox\n\njumps over the lazy dog."},
		{"ft-indent-space.wrj", "The quick\n  brown fox\n    jumps over the\n      lazy dog."},
		{"ft-indent-comment.wrj", "The quick\n  brown fox\n    jumps over the\n      lazy dog."},
		{"ft-embed-comment.wrj", "The quick brown fox\n\n// Not a comment\n\njumps over the lazy dog."},
		{"ft-indent-tab.wrj", "\tThe quick brown fox\n\tjumps over the lazy dog."},
		{"ft-embed-blank-indent-tab.wrj", "\tThe quick brown fox\n\n\tjumps over the lazy dog."},
	} {
		t.Run(test.filename, func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", test.filename))
			if err != nil {
				t.Fatalf("%s", err)
			}

			have := Read(f, "freetext")
			want := Jar{Record{"FREETEXT": []byte(test.want)}}
			compare(t, strconv.Itoa(x), have, want)

			f.Close()
		})
	}
}

func BenchmarkRead(b *testing.B) {
	data, err := ioutil.ReadFile(filepath.Join("testdata", "benchmark.wrj"))
	if err != nil {
		b.Errorf("%s", err)
		return
	}
	r := bytes.NewBuffer(data)

	b.Run(fmt.Sprintf("Read"), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Read(r, "description")
		}
	})
}
