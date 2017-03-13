// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package stats

import (
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/has"

	"log"
	"sync"
)

// players represents a list of all players currently in the game world.
var players = struct {
	list []has.Thing
	sync.Mutex
}{}

// Add adds the specified player to the list of players.
func Add(player has.Thing) {
	players.Lock()
	players.list = append(players.list, player)
	players.Unlock()
}

// Remove removes the specified player from the list of players.
func Remove(player has.Thing) {
	players.Lock()

	for i, p := range players.list {
		if p == player {
			copy(players.list[i:], players.list[i+1:])
			players.list[len(players.list)-1] = nil
			players.list = players.list[:len(players.list)-1]
			break
		}
	}

	// A tiny bit of housekeeping, in case we've had a lot of players recently
	// create a new, smaller capacity player list.
	if len(players.list) == 0 {
		log.Printf("Last one out reclaims the player list: %d slots reclaimed", cap(players.list))
		players.list = make([]has.Thing, 0, 10)
	}

	players.Unlock()
}

func Find(player has.Thing) (found bool) {
	players.Lock()

	for _, p := range players.list {
		if p == player {
			found = true
			break
		}
	}

	players.Unlock()
	return
}

// List returns the names of all players in the player list. The omit parameter
// may be used to specify a player that should be omitted from the list.
func List(omit has.Thing) []string {
	players.Lock()

	list := make([]string, 0, len(players.list))

	for _, player := range players.list {
		if player == omit {
			continue
		}
		list = append(list, attr.FindName(player).Name("Someone"))
	}

	players.Unlock()
	return list
}

// Len returns the length of the player list.
func Len() (l int) {
	players.Lock()
	l = len(players.list)
	players.Unlock()
	return
}
