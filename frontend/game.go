// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/cmd"
	"code.wolfmud.org/WolfMUD.git/has"
	"code.wolfmud.org/WolfMUD.git/message"
	"code.wolfmud.org/WolfMUD.git/stats"
)

// game embeds a frontend instance adding fields and methods specific to
// communicating with the game.
type game struct {
	*frontend
}

// NewGame returns a game with the specified frontend embedded. The returned
// game can be used for processing communication to the actual game.
func NewGame(f *frontend) (g *game) {
	g = &game{frontend: f}
	g.gameInit()
	return
}

// gameInit is used to place the player into the game world. As the game
// backend has it's own output handling we remove the frontend.buf buffer to
// prevent duplicate output. The buffer is restored by gameProcess when the
// player quits the game world.
func (g *game) gameInit() {

	message.ReleaseBuffer(g.buf)
	g.buf = nil
	attr.FindPlayer(g.player).SetPromptStyle(has.StyleBrief)

	i := (*attr.Start)(nil).Pick()
	i.Lock()
	i.Add(g.player)
	stats.Add(g.player)
	i.Unlock()

	cmd.Script(g.player, "$POOF")
	g.nextFunc = g.gameProcess
}

// gameProcess hands input to the game backend for processing while the player
// is in the game. When the player quits the game the frontend.buf buffer is
// restored - see gameInit.
func (g *game) gameProcess() {
	c := cmd.Parse(g.player, string(g.input))
	if c == "QUIT" {
		g.buf = message.AcquireBuffer()
		g.buf.OmitLF(true)
		NewMenu(g.frontend)
	}
}
