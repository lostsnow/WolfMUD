// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"runtime"
)

// version is set at compile time using:
//
//	-ldflags "-X code.wolfmud.org/WolfMUD.git/cmd.version=nnn"
//
// Where nnn is the version. It's format should be the abbreviated git commit
// SHA optionally followed by '-dirty' if the worktree has uncomitted changes.
// e.g. 9a73650 or 9a73650-dirty. This can be handy for debugging user issues.
var version string

func init() {
	if version == "" {
		version = "Unknown"
	}
	version = version + " built with " + runtime.Version()
}

// Syntax: VERSION
func init() {
	AddHandler(Version, "VERSION")
}

func Version(s *state) {
	s.msg.Actor.WriteStrings(version)
	s.ok = true
}
