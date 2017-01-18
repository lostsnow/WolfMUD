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

func compareJars(t *testing.T, j1, j2 Jar) {
j1:
	for x, rec := range j1 {
		for inField, inData := range rec {
			if x > len(j2)-1 {
				t.Errorf("cannot compare output with extra input record %d containing:\n%s", x, printRecord(rec))
				continue j1
			}
			if outData, ok := j2[x][inField]; !ok {
				t.Errorf("record %d field %+q missing from output: %+q\n", x, inField, inData)
			} else {
				if !bytes.Equal(outData, inData) {
					t.Errorf("record %d field %+q data mismatch:\n\tHave: %+q\n\tWant: %+q\n", x, inField, inData, outData)
				}
			}
		}
		if x > len(j2)-1 {
			t.Errorf("cannot compare output with extra input record %d", x)
			continue
		}
		for outField, outData := range j2[x] {
			if _, ok := j1[x][outField]; !ok {
				t.Errorf("record %d field %+q not found in input: %+q\n", x, outField, outData)
			}
		}
	}

	if d := len(j2) - len(j1); d > 0 {
		for x := len(j1); x < len(j2); x++ {
			t.Errorf("extra record %d in output containing:\n%s\n", x, printRecord(j2[x]))
		}
	}
}

func printRecord(r Record) (out []byte) {
	for field, data := range r {
		out = append(out, fmt.Sprintf("%+q: %+q\n", field, data)...)
	}
	return
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
					"FREETEXT": []byte("\tTabbed text two.\n"),
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

		// Server greeting
		{
			`
WolfMUD Copyright 1984-2016 Andrew 'Diddymus' Rolfe

    World
    Of
    Living
    Fantasy

Welcome to WolfMUD!
`,
			Jar{
				Record{"FREETEXT": []byte("\nWolfMUD Copyright 1984-2016 Andrew 'Diddymus' Rolfe\n\n    World\n    Of\n    Living\n    Fantasy\n\nWelcome to WolfMUD!\n")},
			},
		},
	}

	for _, test := range tests {
		b := bytes.NewBufferString(test.data)

		j := Read(b, "freetext")
		compareJars(t, j, test.Jar)
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
