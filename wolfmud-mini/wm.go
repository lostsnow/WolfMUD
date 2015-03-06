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
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

var memprof = flag.Bool("memprof", false, "turn on memory profiling")
var cpuprof = flag.Bool("cpuprof", false, "turn on cpu profiling")

func main() {
	flag.Parse()

	world := attr.Setup()

	// Setup test player
	p := attr.Thing().New(
		attr.Name().New("a player"),
		attr.NewAlias("player"),
		attr.Inventory().New(),
		attr.Locate().New(nil),
	)

	if *memprof {
		runtime.MemProfileRate = 1
		fmt.Println("Memory profileing turned on.")
	}
	if *cpuprof {
		f, _ := os.Create("cpuprof")
		pprof.StartCPUProfile(f)
		fmt.Println("CPU profileing turned on.")
	}

	// Put player into the world
	if i := attr.Exits().Find(world["loc1"]); i != nil {
		i.Place(p)
	}

	// Describe what they can see
	msg, _ := cmd.Parse(p, "LOOK")
	fmt.Printf("%s\n", text.Fold(msg, 80))

	// Main processing loop
	r := bufio.NewReader(os.Stdin)
	fmt.Print(">")
	for i, err := r.ReadString('\n'); err == nil && i != "quit\n"; i, err = r.ReadString('\n') {
		if msg, _ := cmd.Parse(p, i); len(msg) > 0 {
			fmt.Printf("%s\n", text.Fold(msg, 80))
		}
		fmt.Print(">")
	}
	fmt.Println()

	if *memprof {
		f, _ := os.Create("memprof")
		pprof.WriteHeapProfile(f)
		f.Close()
	}
	if *cpuprof {
		pprof.StopCPUProfile()
	}
}
