// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	. "code.wolfmud.org/WolfMUD.git/recordjar"
)

func TestWrite_strings(t *testing.T) {

	longID := strings.Repeat("0123456789", 10)

	for _, test := range []struct {
		data Jar
		want string
	}{
		{Jar{}, ""},                                  // Empty jar
		{Jar{Record{"": []byte("")}}, ""},            // Empty record
		{Jar{Record{"": []byte("d1")}}, ""},          // Empty field name
		{Jar{Record{"F1": []byte("")}}, "F1:\n%%\n"}, // Empty data

		// Single record,  single field
		{Jar{Record{"F1": []byte("d1")}}, "F1: d1\n%%\n"},

		// Single record, single field starting with a non-letter character
		{Jar{Record{"1F": []byte("d1")}}, "1f: d1\n%%\n"},

		// Single record, multiple fields
		{
			Jar{Record{"F1": []byte("d1"), "F2": []byte("d2")}},
			"F1: d1\nF2: d2\n%%\n",
		},
		{
			Jar{Record{"F1": []byte("d1"), "F2": []byte("d2"), "F3": []byte("d3")}},
			"F1: d1\nF2: d2\nF3: d3\n%%\n",
		},

		// Multiple records, single field
		{
			Jar{
				Record{"R1F1": []byte("r1d1")},
				Record{"R2F1": []byte("r2d1")},
			},
			"R1f1: r1d1\n%%\nR2f1: r2d1\n%%\n",
		},
		{
			Jar{
				Record{"R1F1": []byte("r1d1")},
				Record{"R2F1": []byte("r2d1")},
				Record{"R3F1": []byte("r3d1")},
			},
			"R1f1: r1d1\n%%\nR2f1: r2d1\n%%\nR3f1: r3d1\n%%\n",
		},

		// Multiple records, multiple fields
		{
			Jar{
				Record{"R1F1": []byte("r1d1"), "R1F2": []byte("r1d2")},
				Record{"R2F1": []byte("r2d1"), "R2F2": []byte("r2d2")},
			},
			"R1f1: r1d1\nR1f2: r1d2\n%%\nR2f1: r2d1\nR2f2: r2d2\n%%\n",
		},
		{
			Jar{
				Record{"R1F1": []byte("r1d1"), "R1F2": []byte("r1d2")},
				Record{"R2F1": []byte("r2d1"), "R2F2": []byte("r2d2")},
				Record{"R3F1": []byte("r3d1"), "R3F2": []byte("r3d2")},
			},
			"" +
				"R1f1: r1d1\nR1f2: r1d2\n%%\n" +
				"R2f1: r2d1\nR2f2: r2d2\n%%\n" +
				"R3f1: r3d1\nR3f2: r3d2\n%%\n",
		},

		// Multiple fields with continued data lines
		{
			Jar{Record{"F1": []byte("d1\nd2")}},
			"F1: d1\n    d2\n%%\n",
		},
		{
			Jar{Record{"F1": []byte("d1a\nd1b"), "F2": []byte("d2")}},
			"F1: d1a\n    d1b\nF2: d2\n%%\n",
		},
		{
			Jar{Record{"F1": []byte("d1"), "F2": []byte("d2a\nd2b")}},
			"F1: d1\nF2: d2a\n    d2b\n%%\n",
		},

		// Single field with continued data line starting with ': '
		{
			Jar{Record{"F1": []byte("d1\n: d2")}},
			"F1: d1\n  : d2\n%%\n",
		},

		// Single record with free text section (field name in multiple cases)
		{Jar{Record{"ft": []byte("free text.")}}, "free text.\n%%\n"},
		{Jar{Record{"Ft": []byte("Free Text.")}}, "Free Text.\n%%\n"},
		{Jar{Record{"fT": []byte("freE texT.")}}, "freE texT.\n%%\n"},
		{Jar{Record{"FT": []byte("FREE TEXT.")}}, "FREE TEXT.\n%%\n"},

		// Multiple records with free text section
		{
			Jar{
				Record{"FT": []byte("Free text section one.")},
				Record{"FT": []byte("Free text section two.")},
			},
			"Free text section one.\n%%\nFree text section two.\n%%\n",
		},

		// Single record with single field and free text section
		{Jar{Record{"F1": []byte("d1"), "ft": []byte("Free text section.")}},
			"F1: d1\n\nFree text section.\n%%\n",
		},

		// Single record with multiple fields with free text section
		{
			Jar{Record{
				"F1": []byte("d1"),
				"F2": []byte("d2"),
				"ft": []byte("Free text section."),
			}},
			"F1: d1\nF2: d2\n\nFree text section.\n%%\n",
		},
		{
			Jar{Record{
				"F1": []byte("d1"),
				"F2": []byte("d2"),
				"F3": []byte("d3"),
				"ft": []byte("Free text section."),
			}},
			"F1: d1\nF2: d2\nF3: d3\n\nFree text section.\n%%\n",
		},

		// Multiple records with single field and free text section
		{
			Jar{
				Record{"R1F1": []byte("r1d1"), "ft": []byte("Free text section one.")},
				Record{"R2F1": []byte("r2d1"), "ft": []byte("Free text section two.")},
			},
			"" +
				"R1f1: r1d1\n" +
				"\n" +
				"Free text section one.\n" +
				"%%\n" +
				"R2f1: r2d1\n" +
				"\n" +
				"Free text section two.\n" +
				"%%\n",
		},

		// Multiple records with multiple fields and free text section
		{
			Jar{
				Record{
					"R1F1": []byte("r1d1"),
					"R1F2": []byte("r1d2"),
					"ft":   []byte("Free text section one."),
				},
				Record{
					"R2F1": []byte("r2d1"),
					"R2F2": []byte("r2d2"),
					"ft":   []byte("Free text section two."),
				},
			},
			"" +
				"R1f1: r1d1\n" +
				"R1f2: r1d2\n" +
				"\n" +
				"Free text section one.\n" +
				"%%\n" +
				"R2f1: r2d1\n" +
				"R2f2: r2d2\n" +
				"\n" +
				"Free text section two.\n" +
				"%%\n",
		},

		// Long unbreakable field name - wider than the 80 character max width
		{
			Jar{Record{"F" + longID: []byte("d1")}},
			"F" + longID + ": d1\n%%\n",
		},

		// Long unbreakable data value - wider than the 80 character max width
		{
			Jar{Record{"F1": []byte("d" + longID)}},
			"F1: d" + longID + "\n%%\n",
		},

		// Long unbreakable field name and data value - wider than the 80 character
		// max width
		{
			Jar{Record{"F" + longID: []byte("d" + longID)}},
			"F" + longID + ": d" + longID + "\n%%\n",
		},

		//
	} {
		t.Run(fmt.Sprintf("%q", test.want), func(t *testing.T) {
			have := &bytes.Buffer{}
			test.data.Write(have, "ft", nil)
			if have.String() != test.want {
				t.Errorf("have:\n%q\nwant:\n%q", have, test.want)
			}
		})
	}
}

func TestWrite_ordering(t *testing.T) {

	// Seed default random source
	rand.Seed(time.Now().UnixNano())

	want := "" +
		"     Apple: A delecious fruit\n" +
		"       Bat: A flying mammal\n" +
		"Cantaloupe: A type of melon\n" +
		"    Dragon: A mythical creature\n" +
		"  Equation: A mathematical statement\n" +
		"  Flamingo: A pink bird\n" +
		"       Gnu: A type of antelope\n" +
		"%%\n"

	fields := []struct {
		name string
		data string
	}{
		{"gnu", "A type of antelope"},
		{"flamingo", "A pink bird"},
		{"equation", "A mathematical statement"},
		{"dragon", "A mythical creature"},
		{"cantaloupe", "A type of melon"},
		{"bat", "A flying mammal"},
		{"apple", "A delecious fruit"},
	}

	for x := 0; x < 1000; x++ {

		// Assemble Record with fields in a random order
		rec := Record{}
		for _, x := range rand.Perm(len(fields)) {
			rec[fields[x].name] = []byte(fields[x].data)
		}
		test := Jar{rec}

		t.Run(fmt.Sprintf("run %d", x), func(t *testing.T) {
			have := &bytes.Buffer{}
			test.Write(have, "ft", nil)
			if have.String() != want {
				t.Errorf("have:\n%q\nwant:\n%q", have, want)
			}
		})

	}

}

// TestWrite_refolding makes sure that free text sections are unfolded and then
// re-folded when written to format them correctly.
func TestWrite_refolding(t *testing.T) {

	want := `
You are in the corner of the common room in the dragon's breath tavern. A fire
burns merrily in an ornate fireplace, giving comfort to weary travellers. The
fire causes shadows to flicker and dance around the room, changing darkness to
light and back again. To the south the common room continues and east the
common room leads to the tavern entrance.
%%
`

	jar := Jar{Record{"FREETEXT": []byte(`
You are in the corner of the common room in the dragon's
breath tavern. A fire burns merrily in an ornate fireplace,
giving comfort to weary travellers. The fire causes shadows
to flicker and dance around the room, changing darkness to
light and back again. To the south the common room continues
and east the common room leads to the tavern entrance.`)}}

	have := &bytes.Buffer{}
	jar.Write(have, "FREETEXT", nil)
	if have.String() != want {
		t.Errorf("have:\n%q\nwant:\n%q", have, want)
	}

}

func BenchmarkWrite(b *testing.B) {

	location := Record{
		"REF":       []byte("L1"),
		"START":     []byte(""),
		"NAME":      []byte("Fireplace"),
		"ALIASES":   []byte("TAVERN FIREPLACE"),
		"EXITS":     []byte("E→L3 SE→L4 S→L2"),
		"INVENTORY": []byte("L1N1"),
		"FREETEXT": []byte("You are in the corner of the common room in the " +
			"dragon's breath tavern. A fire burns merrily in an ornate fireplace, " +
			"giving comfort to weary travellers. The fire causes shadows to " +
			"flicker and dance around the room, changing darkness to light and " +
			"back again. To the south the common room continues and east the " +
			"common room leads to the tavern entrance."),
	}
	j := Jar{location}

	w := &bytes.Buffer{}

	b.Run(fmt.Sprintf("Write"), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			j.Write(w, "freetext", nil)
			w.Reset()
		}
	})

}
