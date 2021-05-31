// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"code.wolfmud.org/WolfMUD.git/proc"
	"code.wolfmud.org/WolfMUD.git/world"
)

func main() {

	fmt.Println("\n  Welcome to the WolfMini experimental environment!\n")

	world.Load()

	// Setup player
	player := proc.NewThing("Diddymus", "An adventurer, just like you.")
	player.As[proc.Alias] = "PLAYER"
	player.As[proc.Where] = "L1"

	s := proc.NewState(player)
	s.Parse("LOOK")

	var input string
	r := bufio.NewReader(os.Stdin)
	for strings.ToUpper(input) != "QUIT\n" {
		input, _ = r.ReadString('\n')
		s.Parse(input)
	}
}
