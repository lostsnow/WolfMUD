// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/stats"

	"bytes"
)

func (f *frontend) gameInit() {

	f.buf = nil
	attr.FindPlayer(f.player).SetPromptStyle(has.StyleBrief)

	i := (*attr.Start)(nil).Pick()
	i.Lock()
	i.Add(f.player)
	stats.Add(f.player)
	i.Unlock()

	cmd.Parse(f.player, "LOOK")
	f.nextFunc = f.gameProcess
}

func (f *frontend) gameProcess() {
	c := cmd.Parse(f.player, string(f.input))
	if c == "QUIT" {
		f.buf = new(bytes.Buffer)
		f.menuDisplay()
	}
}
