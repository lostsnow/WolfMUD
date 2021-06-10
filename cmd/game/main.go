// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"code.wolfmud.org/WolfMUD.git/proc"
	"code.wolfmud.org/WolfMUD.git/world"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	fmt.Print("\n  Welcome to the WolfMini experimental environment!\n\n")

	world.Load()

	// Setup player
	player := proc.NewThing()
	player.As[proc.Name] = "Diddymus"
	player.As[proc.Description] = "An adventurer, just like you."
	player.As[proc.Where] = proc.WorldStart[rand.Intn(len(proc.WorldStart))]
	player.Any[proc.Alias] = []string{"PLAYER"}

	s := proc.NewState(os.Stdout, player)
	s.Parse("LOOK")

	var input string
	r := bufio.NewReader(os.Stdin)
	for strings.ToUpper(input) != "QUIT\n" {
		input, _ = r.ReadString('\n')
		s.Parse(input)
	}
}
