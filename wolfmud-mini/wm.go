// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package main

import (
	"code.wolfmud.org/WolfMUD-mini.git/attr"
	"code.wolfmud.org/WolfMUD-mini.git/cmd"
	"code.wolfmud.org/WolfMUD-mini.git/text"

	"bufio"
	"fmt"
	"os"
)

func main() {

	world := attr.Setup()

	// Setup test player
	p := attr.Thing().New(
		attr.Name().New("a player"),
		attr.Alias().New("player"),
		attr.Inventory().New(),
		attr.Locate().New(nil),
	)

	// Put player into the world
	if i := attr.FindExit(world["loc1"]); i != nil {
		i.Place(p)
	}

	// Describe what they can see
	fmt.Println(text.Fold(cmd.Parse(p, "LOOK"), 80))

	// Main processing loop
	r := bufio.NewReader(os.Stdin)
	fmt.Print(">")
	for i, err := r.ReadString('\n'); err == nil; i, err = r.ReadString('\n') {
		if o := cmd.Parse(p, i); len(o) > 0 {
			fmt.Println(text.Fold(o, 80))
		}
		fmt.Print(">")
	}
	fmt.Println()
}
