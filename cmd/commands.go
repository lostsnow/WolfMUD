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
	AddHandler(Commands, "COMMANDS")
}

// Width of gutter between columns
const gutter = 2

// BUG(diddymus): Terminal with is hardcoded to be 80 characters wide
func Commands(s *state) {

	cmds := make([]string, len(handlers), len(handlers))

	// Extract keys from handler map
	pos, ommit := 0, 0
	for cmd := range handlers {

		// Ommit empty handler if installed and
		// special commands starting with '#'
		if cmd == "" || cmd[0] == '#' {
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

	// Calculate maximum padding length we will need. We can reslice this for
	// different lengths of padding without allocations
	padding := strings.Repeat(" ", columnWidth)

	// NOTE: cell is not (row * columnCount) + column as we are pivoting the
	// table so that the commands are alphabetical DOWN the rows not across the
	// columns.
	for row := 0; row < rowCount; row++ {
		for column := 0; column < columnCount; column++ {
			cell := (column * rowCount) + row
			if cell < len(cmds) {
				s.msg.actor.WriteJoin(cmds[cell], padding[:columnWidth-len(cmds[cell])])
			}
		}
		s.msg.actor.WriteString("\n")
	}
	s.msg.actor.Truncate(s.msg.actor.Len() - 1)

	s.ok = true
}
