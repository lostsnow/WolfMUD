// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {

	var (
		nbr    = flag.Int("n", 10, "number of bot to launch")
		mins   = flag.Int("t", 1, "number of minutes to run for")
		offset = flag.Int("o", 0, "bot numbering offset")
	)
	flag.Parse()

	log.Printf("Launching %d bots (%d - %d), for %d minutes\n",
		*nbr, *offset, *offset+*nbr-1, *mins)

	// Initialise random number generator with random seed
	rand.Seed(time.Now().UnixNano())

	bots := make([]*Bot, *nbr)
	botg := sync.WaitGroup{}
	botg.Add(*nbr)

	// Create and launch bots
	for x := range bots {
		bots[x] = NewBot(fmt.Sprintf("BOT%d", x+*offset))
		go func(b *Bot) {
			defer botg.Done()
			b.Runner("127.0.0.1", "4001")
		}(bots[x])
		time.Sleep(5 * time.Millisecond) // Don't hammer server too much ;)
		if x%2048 == 2047 {
			log.Printf("Launched: %d bots...", x+1)
			time.Sleep(20 * time.Second) // Don't hammer server too much ;)
		}
	}

	// How long to run for?
	log.Print("Running...")
	time.Sleep(time.Duration(*mins) * time.Minute)

	// Tell all bots to quit and then wait for them to finish
	for _, bot := range bots {
		bot.Quit <- struct{}{}
	}
	botg.Wait()

	log.Print("...finished run.")
}
