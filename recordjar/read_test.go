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
	"testing"

	. "code.wolfmud.org/WolfMUD.git/recordjar"
)

// compare is a helper to compare two Jars j1 and j2.
func compare(t *testing.T, j1, j2 Jar) {

	const (
		extra   = "has extra"
		missing = "is missing"
	)

	t.Helper()
	f := func(reason string) {
		t.Helper()
		for x, r := range j1 {
			if x > len(j2)-1 {
				t.Errorf("%s record %d", reason, x)
				continue
			}
			for field, value := range r {
				if _, ok := j2[x][field]; !ok {
					t.Errorf("record %d - output %s field %q", x, reason, field)
					continue
				}
				if reason == extra && !bytes.Equal(value, j2[x][field]) {
					t.Errorf("record %d, field: %q\nhave: %q\nwant: %q", x, field, j2[x][field], value)
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
		{"\t%%", Jar{}},
		{"%%\t", Jar{}},
		{"\t%%\t", Jar{}},
		{"%%\n%%", Jar{}},
		{"// Comment\n%%", Jar{}},
		{"//Comment\n%%", Jar{}},
		{"  // Comment\n%%", Jar{}},
		{"  //Comment\n%%", Jar{}},
		{"\t// Comment\n%%", Jar{}},
		{"\t//Comment\n%%", Jar{}},

		// Single blank lines
		{"\n", Jar{Record{"FREETEXT": []byte("")}}},
		{"  \n", Jar{Record{"FREETEXT": []byte("")}}},
		{"\t\n", Jar{Record{"FREETEXT": []byte("")}}},
		{"\r\n", Jar{Record{"FREETEXT": []byte("")}}},
		{"  \r\n", Jar{Record{"FREETEXT": []byte("")}}},
		{"\t\r\n", Jar{Record{"FREETEXT": []byte("")}}},

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
		{"F1: d1a\nF2: d2\nF1: d1b", Jar{
			Record{"F1": []byte("d1a d1b"), "F2": []byte("d2")},
		}},

		// Field starting with a non-ASCII letter
		{"1F:d1", Jar{Record{"1F": []byte("d1")}}},
		{"ΔF:d1", Jar{Record{"ΔF": []byte("d1")}}},

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

		// Colon given with no field name
		{":", Jar{Record{"FREETEXT": []byte(":")}}},
		{"  :", Jar{Record{"FREETEXT": []byte("  :")}}},
		{"\t:", Jar{Record{"FREETEXT": []byte("\t:")}}},
		{"F1: d1a\n  : d1b", Jar{Record{"F1": []byte("d1a : d1b")}}},

		// Free text section only
		{"The quick brown fox jumps over the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown fox jumps over the lazy dog."),
				},
			},
		},

		// Free text section over multiple lines
		{"The quick brown\nfox jumps over\nthe lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown\nfox jumps over\nthe lazy dog."),
				},
			},
		},

		// Free text section over multiple indented lines
		{"  The quick brown\n  fox jumps over the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("  The quick brown\n  fox jumps over the lazy dog."),
				},
			},
		},

		// Free text section with leading comment (ignored)
		{"// A comment\nThe quick brown fox jumps over the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown fox jumps over the lazy dog."),
				},
			},
		},

		// Free text section containing comment
		{"The quick brown fox jumps\n// over the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown fox jumps\n// over the lazy dog."),
				},
			},
		},

		// Free text section containing indented comment
		{"The quick brown fox jumps\n  // over the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown fox jumps\n  // over the lazy dog."),
				},
			},
		},

		// Free text section containing blank line
		{"The quick brown fox jumps\n\nover the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown fox jumps\n\nover the lazy dog."),
				},
			},
		},

		// Free text section containing indented separator
		{"The quick brown fox jumps\n  %%\nover the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown fox jumps\n  %%\nover the lazy dog."),
				},
			},
		},

		// Free text section containing non-indented separator
		{"The quick brown fox jumps\n%%\nover the lazy dog.\n%%\n",
			Jar{
				Record{"FREETEXT": []byte("The quick brown fox jumps")},
				Record{"FREETEXT": []byte("over the lazy dog.")},
			},
		},

		// Free text section with leading blank line - should not be mistaken for a
		// separator line and should appear as part of free text section
		{"\nThe quick brown fox jumps over the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("\nThe quick brown fox jumps over the lazy dog."),
				},
			},
		},

		// Free text section with comment and leading blank line
		{"// A comment\n\nThe quick brown fox jumps over the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("\nThe quick brown fox jumps over the lazy dog."),
				},
			},
		},

		// Free text section containing a field, which should be part of the free
		// text section and not seen as a field.
		{"The quick brown fox jumps\nF1: over the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown fox jumps\nF1: over the lazy dog."),
				},
			},
		},

		// Free text section containing an indented  field, which should be part
		// of the free text section and not seen as a field.
		{"The quick brown fox jumps\n  F1: over the lazy dog.\n%%\n",
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown fox jumps\n  F1: over the lazy dog."),
				},
			},
		},

		// Multiple records and free text section + ending separator
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

		// Multiple records and free text section, NO ending separator
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

		// Multiple records and free text section, NO ending separator or new line
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
		t.Run(fmt.Sprintf("#%d_%.20q", x, test.data), func(t *testing.T) {
			have := Read(bytes.NewBufferString(test.data), "freetext", false)
			compare(t, have, test.want)
		})
	}
}

// Test larger data from files being parsed into Jars.
func TestRead_files(t *testing.T) {

	greeting := Record{"FREETEXT": []byte("\nWolfMUD Copyright 1984-2016 Andrew 'Diddymus' Rolfe\n\n    World\n    Of\n    Living\n    Fantasy\n\nWelcome to WolfMUD!\n")}

	location := Record{
		"REF":       []byte("L1"),
		"START":     []byte(""),
		"NAME":      []byte("Fireplace"),
		"ALIASES":   []byte("TAVERN FIREPLACE"),
		"EXITS":     []byte("E→L3 SE→L4 S→L2"),
		"INVENTORY": []byte("L1N1"),
		"FREETEXT":  []byte("You are in the corner of the common room in the dragon's breath tavern. A fire\nburns merrily in an ornate fireplace, giving comfort to weary travellers. The\nfire causes shadows to flicker and dance around the room, changing darkness to\nlight and back again. To the south the common room continues and east the common\nroom leads to the tavern entrance."),
	}

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

			have := Read(f, "freetext", false)
			compare(t, have, test.want)

			f.Close()
		})
	}
}

// Test free text section from files being parsed into Jars. This is easier
// with files than with string literals.
func TestRead_freetext(t *testing.T) {
	for _, test := range []struct {
		filename string
		want     string
	}{
		{"ft-plain.wrj", "The quick\nbrown fox\njumps over\nthe lazy\ndog."},
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

			have := Read(f, "freetext", false)
			want := Jar{Record{"FREETEXT": []byte(test.want)}}
			compare(t, have, want)

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
			_ = Read(r, "description", false)
		}
	})
}
