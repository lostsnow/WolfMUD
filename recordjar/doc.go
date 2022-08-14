// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package recordjar implements the main file format used by WolfMUD. It is
// based on the Record Jar format as described by Eric Raymond in "The Art of
// Unix Programming" [TAOUP, chapter 5].
//
// This implementation differs from the original in a few ways:
//
//   - Unicode is allowed in field names, data and the free text section
//   - Line endings can be CRLF or LF
//   - Comments are lines starting with "//" characters
//   - The free text section is separated from the field section by a blank line
//   - The "%%" record separator may have leading or trailing white space EXCEPT
//     when it follows a free text section - in which case there should be no
//     leading white space.
//
// # Record jar example
//
// This is an example of two records taken from the recordjar defining the
// Zinara zone (see data/zones/zinara.wrj for the full file):
//
//	//
//	// The Dragon's Breath tavern. L1 to L4
//	//
//	    Ref: L1
//	   Type: Start
//	   Name: Fireplace
//	Aliases: TAVERN FIREPLACE
//	  Exits: E→L3 SE→L4 S→L2
//
//	You are in the corner of a common room in the Dragon's Breath tavern.
//	There is a fire burning away merrily in an ornate fireplace giving
//	comfort to weary travellers. Shadows flicker around the room, changing
//	light to darkness and back again. To the south the common room extends
//	and east the common room leads to the tavern entrance.
//	%%
//	    Ref: L2
//	   Type: Start
//	   Name: Common Room
//	Aliases: TAVERN COMMON
//	  Exits: N→L1 NE→L3 E→L4
//
//	You are in a small, cosy common room in the Dragon's Breath tavern.
//	Looking around you see a few chairs and tables for patrons. To the east
//	there is a bar and to the north you can see a merry fireplace burning
//	away.
//
// # Records
//
// Records are separated from each other by a "%%" delimiter on a line by
// itself. Each record consists of an optional field section and an optional
// free text section. If a record has both a field section and a free text
// section they are separated with a single blank line.
//
// # Comment lines
//
// Comments lines are lines that start with a "//" delimiter, optionally with
// leading white space. Comment lines can appear before, after or between field
// lines. Comment lines cannot appear in the free text section as they would be
// taken literally and therefore would become part of the free text data. It is
// valid to have a record that contains only comments.
//
// # Field section
//
// Each record can have a field section. Each field has a name, followed by a
// ":" colon separator, followed by optional data. There should be no white
// space between the field name and the colon separator. The field name may
// contain Unicode but cannot contain any white space. It is a convention to
// use a "-" hyphen where a space would normally be used. Field names are not
// case sensitive, so "Aliases:" and "aliases:" are treated the same. White
// space after the colon separator is not required, but highly recommended to
// aid readability. Field names may be indented with white space - usually to
// align the colon separator to aid readability. A field name may be specified
// more than once, in which case any data will have leading and trailing white
// space removed, the data will then be concatenated together with a single
// space. For example:
//
//	Aliases: apple apples
//	%%
//	Aliases: apple
//	         apples
//	%%
//	Aliases: apple
//	Aliases: apples
//
// All three of these records have a single field named "Aliases" with the data
// value of "apple apples".
//
// The data portion of a field is free format, may contain Unicode and may
// continue over multiple lines. When data is continued over multiple lines it
// may be indented - any leading or trailing white space will be removed and
// the lines concatenated together with a single space. See the second record
// above for an example where the continuation line is indented for alignment
// purposes.
//
// The recordjar format itself does not assign any meaning to field names or
// data values. It is only when used for a specific purpose, for example
// WolfMUD zone files, that specific field names and the content of the data
// have meaning. For example:
//
//	Exits: E→L3 SE→L4 S→L2
//
// It is only within the context of a WolfMUD zone file that the field name
// "Exits" has any meaning. It is also only within that context that the data
// is expected to consist of pairs of values: an exit direction and location
// reference.
//
// It should be noted that care should be taken when using Unicode in field
// names as the Unicode will not be normalized. So, for example, 'Nаme' with a
// Cyrillic 'а' (U+0430) and 'Name' with a Latin 'a' (U+0061) would be treated
// as two different, separate fields.
//
// # Free text section
//
// Each record can have one free text section. If a record also has a field
// section the free text section must appear after the field section, and be
// separated from the field section by a blank line. The separating blank line
// is not considered by of the free text section and will be discarded.
//
// If a record does not contain a field section then a separating blank line
// should not be used.
//
// Within the free text section the following will be preserved:
//
//   - Blank lines
//   - Leading white space
//   - Line breaks before lines with leading white space
//
// An example with line numbers, which are for illustration only and not part
// of the record:
//
//	 1  %%
//	 2  // Server greeting
//	 3
//	 4
//	 5  WolfMUD Copyright 1984-2018 Andrew 'Diddymus' Rolfe
//	 6
//	 7      World
//	 8      Of
//	 9      Living
//	10      Fantasy
//	11
//	12  Welcome to WolfMUD!
//	13
//	14  %%
//
// The example shows a single record consisting of a comment (line 2) and a
// free text section (lines 3-13). The resulting data will have two leading
// blank lines (lines 3-4) and one trailing blank line (line 13). The two blank
// lines on lines 6 and 11 will be preserved. The indenting on lines 7-10 will
// be preserved. The line endings preceding each of the indented lines (7-10)
// will be preserved, causing the indented lines to always start on new lines.
//
// The fee text section may contain Unicode.
//
// Comment lines should not be placed in a free text section as they would be
// taken to be part of the data and not treated as actual comments.
//
// [TAOUP, chapter 5]: http://www.catb.org/esr/writings/taoup/html/ch05s02.html
package recordjar
