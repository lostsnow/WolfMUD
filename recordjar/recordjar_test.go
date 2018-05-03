// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar

import (
	"bytes"
	"fmt"
	"testing"
)

// compare is a helper to compare two Jars j1 and j2. Parameter n can be used
// to identify which jar in a number of jars is being compared.
func compareJars(t *testing.T, id string, j1, j2 Jar) {

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

func TestSimpleRead(t *testing.T) {

	type test struct {
		data string
		Jar
	}

	tests := []test{

		// Empty jars
		{
			"",
			Jar{},
		},
		{
			"%%",
			Jar{},
		},
		{
			"  %%",
			Jar{},
		},
		{
			"%%  ",
			Jar{},
		},
		{
			"  %%  ",
			Jar{},
		},
		{
			"%%\n%%",
			Jar{},
		},
		{
			"// Comment\n%%",
			Jar{},
		},

		// Single field
		{
			"FIELD1: data1",
			Jar{
				Record{
					"FIELD1": []byte("data1"),
				},
			},
		},
		{
			"// Comment\nFIELD1: data1",
			Jar{
				Record{
					"FIELD1": []byte("data1"),
				},
			},
		},
		{
			"// Comment\nFIELD1: data1\n%%",
			Jar{
				Record{
					"FIELD1": []byte("data1"),
				},
			},
		},

		// field prefixed with whitespace
		{
			"// Comment\n  FIELD1: data1\n%%\n",
			Jar{
				Record{
					"FIELD1": []byte("data1"),
				},
			},
		},

		// with trailing line feed
		{
			"// Comment\nFIELD1: data1\n%%\n",
			Jar{
				Record{
					"FIELD1": []byte("data1"),
				},
			},
		},

		// Split field
		{
			"FIELD1: data1a\n        data1b",
			Jar{
				Record{
					"FIELD1": []byte("data1a data1b"),
				},
			},
		},
		// Freetext
		{
			`The quick
brown fox
jumps over the
lazy dog.`,
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown fox jumps over the lazy dog."),
				},
			},
		},
		{
			`The quick brown fox

jumps over the lazy dog.`,
			Jar{
				Record{
					"FREETEXT": []byte("The quick brown fox\n\njumps over the lazy dog."),
				},
			},
		},
		{
			`The quick
  brown fox
    jumps over the
      lazy dog.`,
			Jar{
				Record{
					"FREETEXT": []byte("The quick\n  brown fox\n    jumps over the\n      lazy dog."),
				},
			},
		},
		{
			`The quick
  brown fox
    jumps over the
      lazy dog.
%%`,
			Jar{
				Record{
					"FREETEXT": []byte("The quick\n  brown fox\n    jumps over the\n      lazy dog."),
				},
			},
		},
		{
			`The quick
  brown fox
    jumps over the
      lazy dog.
  %%`,
			Jar{
				Record{
					"FREETEXT": []byte("The quick\n  brown fox\n    jumps over the\n      lazy dog."),
				},
			},
		},
		{
			`The quick
  brown fox
    jumps over the
      lazy dog.
%%  `,
			Jar{
				Record{
					"FREETEXT": []byte("The quick\n  brown fox\n    jumps over the\n      lazy dog."),
				},
			},
		},
		{
			`The quick
  brown fox
    jumps over the
      lazy dog.
  %%  `,
			Jar{
				Record{
					"FREETEXT": []byte("The quick\n  brown fox\n    jumps over the\n      lazy dog."),
				},
			},
		},
		{
			`  // Indented comment
The quick
  brown fox
    jumps over the
      lazy dog.
  %%  `,
			Jar{
				Record{
					"FREETEXT": []byte("The quick\n  brown fox\n    jumps over the\n      lazy dog."),
				},
			},
		},
		{
			`// Comment
FIELD1: data1
FIELD2: data2

Some text one.


%%`,
			Jar{
				Record{
					"FIELD1":   []byte("data1"),
					"FIELD2":   []byte("data2"),
					"FREETEXT": []byte("Some text one.\n\n"),
				},
			},
		},

		// Indented comment
		{
			`  // Comment
  FIELD1: data1
  FIELD2: data2

Some text two.


%%`,
			Jar{
				Record{
					"FIELD1":   []byte("data1"),
					"FIELD2":   []byte("data2"),
					"FREETEXT": []byte("Some text two.\n\n"),
				},
			},
		},

		// Fields with \r\n line endings
		{
			"FIELD1: data1a\r\ndata1b\r\nFIELD2: data2\r\n\r\nSome text three.\r\n\r\n\r\n%%",
			Jar{
				Record{
					"FIELD1":   []byte("data1a data1b"),
					"FIELD2":   []byte("data2"),
					"FREETEXT": []byte("Some text three.\n\n"),
				},
			},
		},
		// Leading tabs
		{
			"\tFIELD1: data1",
			Jar{
				Record{
					"FIELD1": []byte("data1"),
				},
			},
		},
		{
			"\tTabbed text one.",
			Jar{
				Record{
					"FREETEXT": []byte("\tTabbed text one."),
				},
			},
		},
		{
			"\tTabbed text two.\n",
			Jar{
				Record{
					"FREETEXT": []byte("\tTabbed text two."),
				},
			},
		},
		{
			"\tTabbed text three.\n\n%%",
			Jar{
				Record{
					"FREETEXT": []byte("\tTabbed text three.\n"),
				},
			},
		},
		{
			"\tTabbed text four.\n%%\n\tMore tabbing.\n%%",
			Jar{
				Record{"FREETEXT": []byte("\tTabbed text four.")},
				Record{"FREETEXT": []byte("\tMore tabbing.")},
			},
		},
		{
			"\tTabbed text five.\n\nNew line.\n\n%%\n",
			Jar{
				Record{"FREETEXT": []byte("\tTabbed text five.\n\nNew line.\n")},
			},
		},

		// Freetext with an embedded comment
		{
			`
This is a freetext block.

  // This is not a comment but more free text!

This is a freetext block.
` + "\n",
			Jar{
				Record{"FREETEXT": []byte("\nThis is a freetext block.\n\n  // This is not a comment but more free text!\n\nThis is a freetext block.\n")},
			},
		},

		// Server greeting - final "\n" is expected when from a text file
		{
			`
WolfMUD Copyright 1984-2016 Andrew 'Diddymus' Rolfe

    World
    Of
    Living
    Fantasy

Welcome to WolfMUD!
` + "\n",
			Jar{
				Record{"FREETEXT": []byte("\nWolfMUD Copyright 1984-2016 Andrew 'Diddymus' Rolfe\n\n    World\n    Of\n    Living\n    Fantasy\n\nWelcome to WolfMUD!\n")},
			},
		},
	}

	for x, test := range tests {
		b := bytes.NewBufferString(test.data)

		j := Read(b, "freetext")
		compareJars(t, fmt.Sprintf("%d", x), j, test.Jar)
	}

}

var jarData = `
%%
      Ref: ZINARA
     Zone: City of Zinara
   Author: Andrew 'Diddymus' Rolfe

This is the city of Zinara.
%%
      Ref: L1
    Start:
     Name: Fireplace
  Aliases: TAVERN FIREPLACE
    Exits: E→L3 SE→L4 S→L2

You are in the corner of the common room in the dragon's breath tavern. A fire
burns merrily in an ornate fireplace, giving comfort to weary travellers. The
fire causes shadows to flicker and dance around the room, changing darkness to
light and back again. To the south the common room continues and east the common
room leads to the tavern entrance.
%%
      Ref: L2
     Name: Common room
  Aliases: TAVERN COMMON
    Exits: N→L1 NE→L3 E→L4

You are in a small, cosy common room in the dragon's breath tavern. Looking
around you see a few chairs and tables for patrons. To the east you see a bar
and to the north there is the glow of a fire.
%%
    Ref: L3
   Name: Tavern entrance
Aliases: TAVERN ENTRANCE
  Exits: E→L5 S→L4 SW→L2 W→L1

You are in the entryway to the dragon's breath tavern. To the west you see an
inviting fireplace and south an even more inviting bar. Eastward a door leads
out into the street.
%%
      Ref: L4
     Name: Tavern bar
  Aliases: TAVERN BAR
    Exits: N→L3 NW→L1 W→L2

You are at the tavern's very sturdy bar. Behind the bar are shelves stacked with
many bottles in a dizzying array of sizes, shapes and colours. There are also
regular casks of beer, ale, mead, cider and wine behind the bar.
%%
    Ref: L5
   Name: Street between tavern and bakers
Aliases: TAVERN BAKERS STREET
  Exits: N→L14 E→L6 S→L7 W→L3

You are on a well kept cobbled street. Buildings loom up on either side of you.
To the east the smells of a bakery taunt you. To the west the entrance to a
tavern. A sign outside the tavern proclaims it to be the "Dragon's Breath". The
street continues to the north and south.
%%`

func BenchmarkRead(b *testing.B) {
	r := bytes.NewBufferString(jarData)
	b.Run(fmt.Sprintf("Read"), func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Read(r, "description")
		}
	})
}
