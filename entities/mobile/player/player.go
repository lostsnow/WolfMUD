// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package player defines an actual human player in the game.
package player

import (
	"code.wolfmud.org/WolfMUD.git/entities/location"
	"code.wolfmud.org/WolfMUD.git/entities/mobile"
	"code.wolfmud.org/WolfMUD.git/entities/thing"
	"code.wolfmud.org/WolfMUD.git/utils/command"
	"code.wolfmud.org/WolfMUD.git/utils/recordjar"
	"code.wolfmud.org/WolfMUD.git/utils/sender"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
)

// playerCount increments with each player created so we can have unique
// players - created as 'Player n' until we have proper logins.
//
// TODO: Drop playerCount once we have proper logins.
var (
	playerCount = 0
)

// Player is the implementation of a player. Most of the functionallity comes
// from the Mobile type and methods to implement the parser Interface. Apart
// from the parser interface methods Player only contains Player specific code.
type Player struct {
	mobile.Mobile
	sender   sender.Interface
	quitting bool
}

// TODO: loadPlayer currently just generates a player instead of actually
// loading one.
func loadPlayer(sender sender.Interface) (p *Player) {
	playerCount++
	postfix := strconv.Itoa(playerCount)

	r := map[string]string{
		"name":    "Player " + postfix,
		":data:":  "This is player " + postfix,
		"aliases": "Player " + postfix,
	}

	p = &Player{sender: sender}
	p.Unmarshal(r)

	return p
}

func (p *Player) Unmarshal(r recordjar.Record) {
	p.Mobile.Unmarshal(r)
}

// New creates a new Player and returns a reference to it. The player is put
// into the world at a random starting location and the location is described.
func New(sender sender.Interface) (p *Player) {
	p = loadPlayer(sender)
	p.add(location.GetStart())
	return p
}

// IsQuitting returns true if the player is trying to quit otherwise false. It
// implements part of the parser.Interface.
func (p *Player) IsQuitting() bool {
	return p.quitting
}

// Destroy should cleanly shutdown the Parser when called. It implements part
// of the parser.Interface.
func (p *Player) Destroy() {

	// execute p.remove until successful ... looks weird ;)
	for !p.remove() {
	}

	p.sender = nil
}

// add places a player in the world safely and announces their arrival.  We
// manually build and parse the 'LOOK' command to avoid deadlocking - adding
// the player locks the location as does a normal p.Parse('LOOK'). We could add
// the player and then parse but that would require obtaining the lock twice.
func (p *Player) add(l location.Interface) {
	l.Lock()
	defer l.Unlock()

	l.Add(p)
	PlayerList.Add(p)

	cmd := command.New(p, "LOOK")
	p.Process(cmd)

	if !l.Crowded() {
		cmd.Broadcast([]thing.Interface{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.Name())
	}

	cmd.Flush()
}

// remove extracts a player from the world cleanly and announces their
// departure.
func (p *Player) remove() (removed bool) {
	l := p.Locate()
	l.Lock()
	defer l.Unlock()

	if l.IsAlso(p.Locate()) {

		// Quitting or involuntary disconnection?
		if p.IsQuitting() {
			p.Locate().Broadcast([]thing.Interface{p}, "%s gives a strangled cry of 'Bye Bye', and then slowly fades away and is gone.", p.Name())
		} else {
			PlayerList.Broadcast(nil, "AAAaaarrrggghhh!!!\nA scream is heard across the land as %s is unceremoniously extracted from the world.", p.Name())
		}

		l.Remove(p)
		PlayerList.Remove(p)
		removed = true
	}

	return
}

// dropInventory drops everything the player is carrying.
func (p *Player) dropInventory(cmd *command.Command) {
	for _, o := range p.Inventory.List() {
		if c, ok := o.(command.Interface); ok {
			if aliases := o.Aliases(); len(aliases) > 0 {
				cmd.New("DROP " + o.Aliases()[0])
				c.Process(cmd)
			}
		}
	}
}

// Parse takes a string and begins the delegation to potential processors. To
// avoid deadlocks, inconsistencies, race conditions and other unmentionables
// we lock the location of the player. However there is a race condition
// between getting the player's location and locking it - they may have moved
// in-between. We therefore get and lock their current location then check it's
// still their current location. If it is not we unlock and try again.
//
// If a command effects more than one location we have to release the current
// lock on the location and relock the locations in Unique Id order before
// trying again. Always locking in a consistent order greatly helps in avoiding
// deadlocks.
//
// MOST of the time we are only interested in a few things: The current player,
// it's location, items at the location, mobiles at the location. We can
// therefore avoid complex fine grained locking on each individual Thing and
// just lock on the whole location. This does mean if there are a LOT of things
// happening in one specific location we will not have as much parallelism as we
// would like.
//
// TODO: If there many clients trying to connect at once - say 250+ simultaneous
// clients connecting - then the starting location becomes a bit of a bottle
// neck (at 1,000+ simultaneous clients connecting is a pain - but once
// connected things smooth out and become playable again). Adding more starting
// locations help to spread the bottle neck. Note that this is just an issue
// with the initial connection and multiple clients all trying to grab the start
// location lock!
func (p *Player) Parse(input string) {

	cmd := command.New(p, input)
	cmd.AddLock(p.Locate())
	cmd.LocksModified()

	// Another funky looking for loop :)
	for p.parseStage2(cmd) {
	}

	cmd.Flush()
}

// parseStage2 is called by Parse to take advantage of defer unwinding. By
// splitting the parsing we can easily obtain the locks we want and defer the
// unlocking. This makes both Parse and parseStage2 very simple.
func (p *Player) parseStage2(cmd *command.Command) (retry bool) {
	for _, l := range cmd.Locks {
		l.Lock()
		defer l.Unlock()
	}

	// If player moved before we locked we need to retry
	if !cmd.CanLock(p.Locate()) {
		return true
	}

	handled := p.Process(cmd)
	retry = cmd.LocksModified()

	if !handled && !retry {
		cmd.Respond("Eh?")
	}

	return
}

// NOTE: We should never have a nil sender as it's deallocated only after the
// player is extracted from the world.
func (p *Player) Respond(format string, any ...interface{}) {
	p.sender.Send(format, any...)
}

// Broadcast implements the broadcaster interface and broadcasts to the
// player's current location.
func (p *Player) Broadcast(omit []thing.Interface, format string, any ...interface{}) {
	p.Locate().Broadcast(omit, format, any...)
}

// Process implements the command.Interface to handle player specific commands.
// It also delegates to mobile.Process if the player can't handle the command
// which also does most of the delegating to get commands processed. . As a
// last resort we see if PlayerList can handle the command. PlayerList can't be
// handled by Mobile with everything else as it causes a cyclic import and goes
// BOOM!
func (p *Player) Process(cmd *command.Command) (handled bool) {

	switch cmd.Verb {
	case "CPUSTOP":
		handled = p.cpustop(cmd)
	case "CPUSTART":
		handled = p.cpustart(cmd)
	case "MEMPROF":
		handled = p.memprof(cmd)
	case "QUIT":
		handled = p.quit(cmd)
	case "SNEEZE":
		handled = p.sneeze(cmd)
	}

	if !handled {
		handled = p.Mobile.Process(cmd)
	}

	if !handled {
		handled = PlayerList.Process(cmd)
	}

	return
}

// cpustart implement the 'CPUSTART' command and starts CPU profiling.
//
// TODO: Remove - for debugging only
func (p *Player) cpustart(cmd *command.Command) (handled bool) {
	f, err := os.Create("cpuprof")
	if err != nil {
		cmd.Respond("CPU Profile Not Started: %s", err)
		return false
	}
	pprof.StartCPUProfile(f)

	cmd.Respond("CPU profile started")
	return true
}

// cpustop implements the 'CPUSTOP' command, stops CPU profiling and writes the
// profile to cpuprofile in the bin directory.
//
// TODO: Remove - for debugging only
func (p *Player) cpustop(cmd *command.Command) (handled bool) {
	pprof.StopCPUProfile()
	cmd.Respond("CPU profile stopped")
	return true
}

// memprof implements the 'MEMPROF' command and writes out a memprofile.
//
// NOTE: Need to change the value of MemProfileRate in server.go
// TODO: Remove - for debugging only
func (p *Player) memprof(cmd *command.Command) (handled bool) {
	f, err := os.Create("memprof")
	if err != nil {
		cmd.Respond("Memory Profile Not Dumped: %s", err)
		return false
	}
	pprof.WriteHeapProfile(f)
	f.Close()

	cmd.Respond("Memory profile dumped")
	return true
}

// quit implements the 'QUIT' command.
//
// TODO: Document exact effect when finalised and Destroy etc cleaned
// up/possibly removed.
func (p *Player) quit(cmd *command.Command) (handled bool) {
	p.dropInventory(cmd)
	cmd.Respond("\n[YELLOW]Bye Bye[WHITE]\n")
	p.quitting = true
	log.Printf("%s is quitting", p.Name())
	p.sender.Prompt("")
	return true
}

// sneeze implements the 'SNEEZE' command.
//
// TODO: Remove - for debugging responders and broadcasters
func (p *Player) sneeze(cmd *command.Command) (handled bool) {
	cmd.Respond("You sneeze. Aaahhhccchhhooo!")
	cmd.Broadcast([]thing.Interface{p}, "You see %s sneeze.", cmd.Issuer.Name())
	PlayerList.Broadcast(p.Locate().List(), "You hear a loud sneeze.")
	return true
}
