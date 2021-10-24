// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.
package client

import (
	"math/rand"
	"strings"

	"code.wolfmud.org/WolfMUD.git/core"
)

func (c *client) enterWorld() {
	core.BWL.Lock()
	c.Ref[core.Where] = core.WorldStart[rand.Intn(len(core.WorldStart))]
	c.Ref[core.Where].Who[c.uid] = c.Thing
	core.BWL.Unlock()
}

func (c *client) createPlayer() {
	c.Is = c.Is | core.Player
	c.As[core.UName] = c.As[core.Name]
	c.As[core.TheName] = c.As[core.Name]
	c.As[core.UTheName] = c.As[core.Name]
	c.As[core.Description] = "An adventurer, just like you."
	c.As[core.DynamicAlias] = "PLAYER"
	c.Any[core.Alias] = []string{"PLAYER", strings.ToUpper(c.As[core.Name])}
	c.Any[core.Body] = []string{
		"HEAD",
		"FACE", "EAR", "EYE", "NOSE", "EYE", "EAR",
		"MOUTH", "UPPER_LIP", "LOWER_LIP",
		"NECK",
		"SHOULDER", "UPPER_ARM", "ELBOW", "LOWER_ARM", "WRIST",
		"HAND", "FINGER", "FINGER", "FINGER", "FINGER", "THUMB",
		"SHOULDER", "UPPER_ARM", "ELBOW", "LOWER_ARM", "WRIST",
		"HAND", "FINGER", "FINGER", "FINGER", "FINGER", "THUMB",
		"BACK", "CHEST",
		"WAIST", "PELVIS",
		"UPPER_LEG", "KNEE", "LOWER_LEG", "ANKLE", "FOOT",
		"UPPER_LEG", "KNEE", "LOWER_LEG", "ANKLE", "FOOT",
	}
}
