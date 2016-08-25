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

func (d *Driver) gameSetup() {

	d.buf = nil
	attr.FindPlayer(d.player).SetPromptStyle(has.StyleBrief)

	i := (*attr.Start)(nil).Pick()
	i.Lock()
	i.Add(d.player)
	stats.Add(d.player)
	i.Unlock()

	cmd.Parse(d.player, "LOOK")
	d.nextFunc = d.gameRun
}

func (d *Driver) gameRun() {
	c := cmd.Parse(d.player, string(d.input))
	if c == "QUIT" {
		d.buf = new(bytes.Buffer)
		d.menuDisplay()
	}
}
