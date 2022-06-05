// Copyright 2022 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package term provides functions for controlling the terminal using ANSI
// escape sequences.
//
// The current functionality has been tested with the Linux TELNET client,
// Windows TELNET client and Putty.
package term

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"

	"code.wolfmud.org/WolfMUD.git/text"
)

const (
	msg = "" +
		"     Will try to determine terminal size,\r\n" +
		"     otherwise 80x25 will be assumed.\r\n" +
		"     Hit enter to continue..."

	ED  = text.CSI + "2J" // Erase in display (whole screen)
	EL  = text.CSI + "0K" // Erase line (cursor to end)
	DSR = text.CSI + "6n" // Device Status Report

	DECSC = text.ESC + "7" // DEC Save Cursor
	DECRC = text.ESC + "8" // DEC Restore Cursor
)

var (
	// Position Cursor (row, column)
	CUP = func(r, c int) string {
		return text.CSI + strconv.Itoa(r) + ";" + strconv.Itoa(c) + "H"
	}

	// DEC Set Scroll Region (top, bottom)
	DECSTBM = func(t, b int) string {
		return text.CSI + strconv.Itoa(t) + ";" + strconv.Itoa(b) + "r"
	}
)

// GetSize attempts to retrieve the current terminal's width and height
// (columns and lines). If the size cannot be determined a default of 80x25
// will be returned.
func GetSize(rw io.ReadWriter) (width, height int) {

	rw.Write([]byte(
		ED + CUP(3, 1) + msg + CUP(255, 255) + text.Black + text.BGBlack + DSR,
	))

	var err error
	w, h := 80, 25
	if size := filterSize(rw, '\n', 7); len(size) > 0 {
		hb, wb, found := bytes.Cut(size, []byte(";"))
		if found {
			if w, err = strconv.Atoi(string(wb)); err != nil || len(wb) == 0 {
				w = 80
			}
			if h, err = strconv.Atoi(string(hb)); err != nil || len(hb) == 0 {
				h = 25
			}
		}
	}

	rw.Write([]byte(text.Reset + ED + CUP(h, 1) + DECSC))

	return w, h
}

// Setup returns a []byte to display the initial status and input areas.
func Setup(width, height int) []byte {
	d := Status(height, width)
	d = append(d, []byte(fmt.Sprintf(" Terminal: %dx%d", width, height))...)
	d = append(d, Input(height)...)
	return d
}

// Output returns a []byte that should prefix any data writen to the output
// area.
func Output(height int) []byte {
	return []byte(
		DECSC + DECSTBM(1, height-3) + CUP(height-3, 1),
	)
}

// Status returns a []byte that should prefix any data writen to the status
// area.
func Status(height, width int) []byte {
	return []byte(CUP(height-2, 1) + text.BGBlue + text.White + EL)
}

// Input returns a []byte that should prefix any data writen to the input area.
func Input(height int) []byte {
	return []byte(DECSTBM(height-1, height) + DECRC + text.Reset + text.Prompt)
}

// ftab is a state table for filterSize.
var ftab = []struct {
	keep   bool
	repeat bool
	lo     byte
	hi     byte
}{
	{false, false, 0x1b, 0x1b},
	{false, false, '[', '['},
	{true, true, '0', '9'},
	{true, false, ';', ';'},
	{true, true, '0', '9'},
	{false, false, 'R', 'R'},
}

// filterSize filters out the response to a "Device Status Report (CSI 6n)"
// request. When the request is sent the client may not send the response
// immediately, or may respond out of sequence. filterSize reads the incoming
// data and filters out just the "row;column" of the response as a []byte. If
// the response is not found or there is an error an empty slice is returned.
// The size is returned as []byte("r;c") where 'r' is the rows and 'c' the
// columns, for example: "25;80".
func filterSize(r io.Reader, term byte, limit int) []byte {
	read := []byte{}
	stg, pos := make([]byte, 0, limit), 0
	b := bufio.NewReader(r)
	defer b.Reset(r)
	for d, err := b.ReadByte(); err == nil; d, err = b.ReadByte() {
		read = append(read, d)
	retry:
		if d == term {
			return stg[:0]
		}
		if d < ftab[pos].lo || ftab[pos].hi < d {
			if ftab[pos].repeat {
				pos++
				goto retry
			}
			stg, pos = stg[:0], 0
			continue
		}
		if ftab[pos].keep {
			if stg = append(stg, d); len(stg) > limit {
				return stg[:0]
			}
			if ftab[pos].repeat {
				continue
			}
		}
		if pos++; pos == len(ftab) {
			return stg
		}
	}
	return stg[:0]
}
