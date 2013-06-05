// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar

import (
	"strings"
	"testing"
)

var testSubjects = []struct {
	input  string
	output RecordJar
}{

	// A simple test - one record one header
	{"header: A simple test.",
		[]Record{
			{"header": "A simple test."},
		},
	},

	// Test with one record and two headers
	{`header1: Two header test - header 1.
		header2: Two header test - header 2.`,
		[]Record{
			{
				"header1": "Two header test - header 1.",
				"header2": "Two header test - header 2.",
			},
		},
	},

	// Test with one record and two headers the same
	{`header: Two header test - header 1.
		header: Two header test - header 2.`,
		[]Record{
			{
				"header": "Two header test - header 1. Two header test - header 2.",
			},
		},
	},

	// Test with one header split over two lines
	{`header: A longer
						continuation line test`,
		[]Record{
			{
				"header": "A longer continuation line test",
			},
		},
	},

	// Empty test
	{},

	// Comment only
	{
		"// A comment",
		[]Record{},
	},

	// Separator only
	{
		"%%",
		[]Record{},
	},

	// Multiple separators only
	{
		`%%
%%
%%`,
		[]Record{},
	},

	// Comment and separator only
	{
		"// Comment only!\n%%",
		[]Record{},
	},

	// Test with one record, multiple + split headers
	{`Abstract: longer test
		Description: A longer test with a data segment

This is the data segment for the longer test. We
should expect this to pass!`,
		[]Record{
			{
				"abstract":    "longer test",
				"description": "A longer test with a data segment",
				":data:":      "This is the data segment for the longer test. We should expect this to pass!",
			},
		},
	},

	// Multiple records with one header the same
	{`header: record one
%%
header: record two`,
		[]Record{
			{
				"header": "record one",
			},
			{
				"header": "record two",
			},
		},
	},

	// Multiple records with multiple separators, one header the same per record
	{`header: record one
%%
%%
%%
header: record two`,
		[]Record{
			{
				"header": "record one",
			},
			{
				"header": "record two",
			},
		},
	},

	// Typical sample of multiple location records
	{`Ref: L1
	 Type: Start
	 Name: Fireplace
Aliases: TAVERN FIREPLACE
	Exits: E→L3 SE→L4 S→L2

You are in the corner of a common room in the Dragon's Breath tavern.
There is a fire burning away merrily in an ornate fireplace giving
comfort to weary travellers. Shadows flicker around the room, changing
light to darkness and back again. To the south the common room extends
and east the common room leads to the tavern entrance.
%%
		Ref: L2
	 Type: Start
	 Name: Common Room
Aliases: TAVERN COMMON
	Exits: N→L1 NE→L3 E→L4

You are in a small, cosy common room in the Dragon's Breath tavern.
Looking around you see a few chairs and tables for patrons. To the east
there is a bar and to the north you can see a merry fireplace burning
away.`,

		[]Record{
			{
				"ref":     "L1",
				"type":    "Start",
				"name":    "Fireplace",
				"aliases": "TAVERN FIREPLACE",
				"exits":   "E→L3 SE→L4 S→L2",
				":data:":  "You are in the corner of a common room in the Dragon's Breath tavern. There is a fire burning away merrily in an ornate fireplace giving comfort to weary travellers. Shadows flicker around the room, changing light to darkness and back again. To the south the common room extends and east the common room leads to the tavern entrance.",
			},
			{
				"ref":     "L2",
				"type":    "Start",
				"name":    "Common Room",
				"aliases": "TAVERN COMMON",
				"exits":   "N→L1 NE→L3 E→L4",
				":data:":  "You are in a small, cosy common room in the Dragon's Breath tavern. Looking around you see a few chairs and tables for patrons. To the east there is a bar and to the north you can see a merry fireplace burning away.",
			},
		},
	},

	// Unicode data testing - If I've got a mistake here or
	// mangled something in the text please let me know :(
	{
		`English: Hello World
			Arabic: مرحبا العالم
		 Chinese: 你好世界
			Hebrew: שלום עולם
		Japanese: こんにちは世界
		  Korean: 안녕하세요 세계
			 Hindi: नमस्ते दुनिया
			  Urdu: ہیلو دنیا
		    Thai: สวัสดีชาวโลก`,

		[]Record{
			{
				"english":  "Hello World",
				"arabic":   "مرحبا العالم",
				"chinese":  "你好世界",
				"hebrew":   "שלום עולם",
				"japanese": "こんにちは世界",
				"korean":   "안녕하세요 세계",
				"hindi":    "नमस्ते दुनिया",
				"urdu":     "ہیلو دنیا",
				"thai":     "สวัสดีชาวโลก",
			},
		},
	},

	// Unicode header testing - If I've got a mistake here or
	// mangled something in the text please let me know :(
	{
		`peace: English
			سلام:  Arabic
		 和平:  Chinese
			שלום:	Hebrew
			平和:	Japanese
			평화:	Korean
			 ﺺﻠﺣ: Urdu
		สันติภาพ:	Thai`,

		[]Record{
			{
				"peace":    "English",
				"سلام":     "Arabic",
				"和平":       "Chinese",
				"שלום":     "Hebrew",
				"平和":       "Japanese",
				"평화":       "Korean",
				"ﺺﻠﺣ":      "Urdu",
				"สันติภาพ": "Thai",
			},
		},
	},

	// Test with one record and two headers + DOS line endings
	{"header1: Two header test - header 1.\r\nheader2: Two header test - header 2.\r\n\r\nThis is the description\r\nwith a line break.",
		[]Record{
			{
				"header1": "Two header test - header 1.",
				"header2": "Two header test - header 2.",
				":data:":  "This is the description with a line break.",
			},
		},
	},
}

func TestRead(t *testing.T) {

	for i, s := range testSubjects {
		r := strings.NewReader(s.input)
		rj, err := Read(r)

		if err != nil {
			t.Errorf("Read error: Case %d, error %q", i, err)
		}

		// Do we have enough Records in the RecordJar?
		have := len(rj)
		want := len(s.output)
		if have != want {
			t.Errorf("Invalid record count: Case %d, have %d wanted %d", i, have, want)
			break
		}

		// For each Record in the RecordJar ...
		for j, r := range rj {

			// ... do we have enough headers?
			have := len(r)
			want := len(s.output[j])
			if have != want {
				t.Errorf("Invalid header count: Case %d, have %d wanted %d", i, have, want)
				break
			}

			// ... do we have any unexpected headers ...
			for header, data := range r {
				have := data
				want, ok := s.output[j][header]
				if !ok {
					t.Errorf("Unexpected header: Case %d, have %q", i, header)
				}

				// ... do we have the expected data?
				if have != want {
					t.Errorf("Header corrupt: Case %d, have %q wanted %q", i, have, want)
				}
			}

			// ... any missing headers?
			for want := range s.output[j] {
				_, ok := r[want]
				if !ok {
					t.Errorf("Missing header: Case %d, wanted %q", i, want)
				}
			}
		}

	}
}
