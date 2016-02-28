// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package recordjar implements the main file format used by WolfMUD.  It is
// based on a combination of RFC5322 and the Record Jar format as described by
// Eric Raymond in "The Art of Unix Programming", chapter 5:
//
//	http://www.catb.org/esr/writings/taoup/html/ch05s02.html
//
// It is not an actual implementation of the RFC5322 format just based on it:
//
//	- Unicode is allowed in field names and data
//	- Whitespace handling is more lenient
//	- Line endings can be CRLF or LF
//	- Comments are lines starting with '//' characters
//	- Multiple records are separated by the '%%' sequence
//	- If a record contains only a free text block a blank line does not need to
//		preceed it.
//
// White space may proceed field names, comments or record separators. Leading
// whitespace and blank lines will be preserved in the free text section.
//
// Here is a simple example of two starting locations:
//
//		//
//		// The Dragon's Breath tavern. L1 to L4
//		//
//				Ref: L1
//			 Type: Start
//			 Name: Fireplace
//		Aliases: TAVERN FIREPLACE
//			Exits: E→L3 SE→L4 S→L2
//
//		You are in the corner of a common room in the Dragon's Breath tavern.
//		There is a fire burning away merrily in an ornate fireplace giving
//		comfort to weary travellers. Shadows flicker around the room, changing
//		light to darkness and back again. To the south the common room extends
//		and east the common room leads to the tavern entrance.
//		%%
//				Ref: L2
//			 Type: Start
//			 Name: Common Room
//		Aliases: TAVERN COMMON
//			Exits: N→L1 NE→L3 E→L4
//
//		You are in a small, cosy common room in the Dragon's Breath tavern.
//		Looking around you see a few chairs and tables for patrons. To the east
//		there is a bar and to the north you can see a merry fireplace burning
//		away.
//
//
// When this example is read you would have a Jar which is a slice of Records -
// in this case two of them.
package recordjar
