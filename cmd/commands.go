// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"sort"
	"strings"
)

// Syntax: COMANDS
func init() {
	addHandler(commands{}, "COMMANDS")
}

// Width of gutter between columns
const gutter = 2

type commands cmd

// BUG(diddymus): Terminal with is hardcoded to be 80 characters wide
func (commands) process(s *state) {

	cmds := make([]string, len(handlers), len(handlers))

	// Extract keys from handler map
	pos, ommit := 0, 0
	for cmd := range handlers {

		// Ommit empty handler if installed and special commands starting with '#'
		// and scripting commands starting with '$'
		if cmd == "" || cmd[0] == '#' || cmd[0] == '$' {
			ommit++
			continue
		}

		cmds[pos] = cmd
		pos++
	}

	// Reslice to remove omitted slots
	cmds = cmds[0 : len(cmds)-ommit]

	// Find longest key extracted
	maxWidth := 0
	for _, cmd := range cmds {
		if l := len(cmd); l > maxWidth {
			maxWidth = l
		}
	}

	sort.Strings(cmds)

	var (
		columnWidth = maxWidth + gutter
		columnCount = 80 / columnWidth
		rowCount    = (len(cmds) / columnCount)
	)

	// If we have a partial row we need to account for it
	if len(cmds) > rowCount*columnCount {
		rowCount++
	}

	// NOTE: cell is not (row * columnCount) + column as we are pivoting the
	// table so that the commands are alphabetical DOWN the rows not across the
	// columns.
	for row := 0; row < rowCount; row++ {
		line := []byte{}
		for column := 0; column < columnCount; column++ {
			cell := (column * rowCount) + row
			if cell < len(cmds) {
				line = append(line, cmds[cell]...)
				line = append(line, strings.Repeat("␠", columnWidth-len(cmds[cell]))...)
			}
		}
		s.msg.Actor.Send(string(line))
	}

	s.ok = true
}
