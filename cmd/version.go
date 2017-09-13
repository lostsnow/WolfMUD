// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"runtime"
)

// commit is set at compile time using:
//
//	-ldflags "-X code.wolfmud.org/WolfMUD.git/cmd.commit=id"
//
// Where id is the git commit returned by git describe HEAD when compiling
// optionally followed by '-dirty' if the worktree has uncomitted changes. For
// example v0.0.7-2-g89fdad9 or v0.0.7-2-g89fdad9-dirty. This can be handy for
// debugging user issues.
var commit string

func init() {
	if commit == "" {
		commit = "Unknown"
	}
	commit = commit + " built with " + runtime.Version()
}

// Syntax: VERSION
func init() {
	AddHandler(version{}, "VERSION")
}

type version cmd

func (version) process(s *state) {
	s.msg.Actor.SendInfo(commit)
	s.ok = true
}
