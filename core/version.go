// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package core

// commit is overridden at compile time using:
//
//	-ldflags "-X code.wolfmud.org/WolfMUD.git/core.commit=id"
//
// Where id is the git commit returned by git describe HEAD when compiling,
// optionally followed by '-dirty' if the worktree has uncomitted changes. For
// example v0.0.7-2-g89fdad9 or v0.0.7-2-g89fdad9-dirty.
var commit = "unknown"
