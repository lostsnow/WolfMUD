// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package player defines an actual human player in the game.
package player

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"wolfmud.org/entities/location"
	"wolfmud.org/entities/mobile"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/command"
	"wolfmud.org/utils/sender"
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
	*mobile.Mobile
	sender   sender.Interface
	name     string
	quitting bool
}

// New creates a new Player and returns a reference to it.
func New(sender sender.Interface, l location.Interface) *Player {

	playerCount++
	postfix := strconv.Itoa(playerCount)

	name := "Player " + postfix
	alias := []string{"PLAYER" + postfix}
	description := "This is Player " + postfix + "."

	p := &Player{
		Mobile: mobile.New(name, alias, description),
		sender: sender,
	}
	p.name = p.Name()

	// Put player into the world
	p.add(l)

	// Describe starting location.
	// NOTE: Don't put this into add() otherwise we deadlock because both do locking!
	p.Parse("LOOK")

	log.Printf("Player %d created: %s\n", p.UniqueId(), p.Name())
	runtime.SetFinalizer(p, final)

	return p
}

// final is used for debugging to make sure the GC is cleaning up
func final(p *Player) {
	log.Printf("+++ %s finalized +++\n", p.name)
}

// IsQuitting returns true if the player is trying to quit otherwise false. It
// implements part of the parser.Interface.
func (p *Player) IsQuitting() bool {
	p.Lock()
	defer p.Unlock()
	return p.quitting
}

// Destroy should cleanly shutdown the Parser when called. It implements part
// of the parser.Interface.
func (p *Player) Destroy() {

	name := p.Name()

	log.Printf("Destroy: %s\n", name)

	if p.IsQuitting() {
		log.Printf("%s is quitting @ %s", name, p.Locate().Name())
		p.Locate().Broadcast(nil, "%s gives a strangled cry of 'Bye Bye', and then slowly fades away and is gone.", name)
	}

	// execute p.remove until successful ... looks weird ;)
	for !p.remove() {
	}

	// Involuntary disconnection?
	if !p.IsQuitting() {
		PlayerList.Broadcast(nil, "AAAaaarrrggghhh!!!\nA scream is heard across the land as %s is unceremoniously extracted from the world.", name)
	}

	p.sender = nil
	p.Mobile = nil

	log.Printf("Destroyed: %s\n", name)
}

// add places a player in the world safely.
func (p *Player) add(l location.Interface) {

	l.Lock()
	defer l.Unlock()

	l.Add(p)
	l.Broadcast(
		[]thing.Interface{p},
		"There is a puff of smoke and %s appears spluttering and coughing.",
		p.Name(),
	)

	PlayerList.Add(p)
}

// remove extracts a player from the world cleanly.
func (p *Player) remove() (removed bool) {
	l := p.Locate()
	l.Lock()
	defer l.Unlock()
	if l.IsAlso(p.Locate()) {
		p.Locate().Remove(p)
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
// trying again. Locking in a consistent order avoids deadlocks.
//
// MOST of the time we are only interested in a few things: The current player,
// it's location, items at the location, mobiles at the location. We can
// therefore avoid complex fine grained locking on each individual Thing and
// just lock on the whole location. This does mean if there are a LOT of things
// happening in one specific location we will not have as much parallelism as we
// would like.
//
// TODO: If there many clients trying to connect at once - say 250+ simultaneous
// clients connecting - then the starting location becomes a bottle neck.
// Adding more starting locations help spread the bottle neck. What might be an
// idea is to make parseStage2 smarter and not release locks if they are lower
// than the locks we now require. For example if we have the ID 100 for the
// current location and we want to move EAST to location with ID 105 can we just
// lock 105 instead of unlocking 100 and then locking 100 and 105? Would this
// give any advantage or disadvantage to players having to unlock and relock
// as opposed to those who just do the extra lock? I don't think this would
// deadlock as it would be similar to the current locking but with a slightly
// longer delay between some locks being taken.
func (p *Player) Parse(input string) {

	cmd := command.New(p, input)
	cmd.AddLock(p.Locate())

	for retry := cmd.LocksModified(); retry; {
		retry = p.parseStage2(cmd)
	}

	cmd.Flush()

}

func (p *Player) parseStage2(cmd *command.Command) (retry bool) {
	for _, l := range cmd.Locks {
		l.Lock()
		defer l.Unlock()
	}
	if cmd.CanLock(p.Locate()) {
		handled := p.Process(cmd)
		retry = cmd.LocksModified()
		if handled == false && !retry {
			cmd.Respond("Eh?")
		}
	} else {
		retry = true
	}
	return
}

func (p *Player) Respond(format string, any ...interface{}) {
	if c := p.sender; c != nil {
		c.Send(format, any...)
		runtime.Gosched()
	} else {
		log.Printf("Respond: %s is a Zombie\n", p.name)
	}
}

func (p *Player) Broadcast(omit []thing.Interface, format string, any ...interface{}) {
	p.Locate().Broadcast(omit, format, any...)
}

func (p *Player) Process(cmd *command.Command) (handled bool) {

	switch cmd.Verb {
	default:
		handled = p.Mobile.Process(cmd)
	case "MEMPROF":
		handled = p.memprof(cmd)
	case "QUIT":
		handled = p.quit(cmd)
	case "SNEEZE":
		handled = p.sneeze(cmd)
	case "WHO":
		handled = p.who(cmd)
	}

	return
}

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

func (p *Player) quit(cmd *command.Command) (handled bool) {
	p.Lock()
	defer p.Unlock()
	p.dropInventory(cmd)
	cmd.Respond("\n[YELLOW]Bye Bye[WHITE]\n")
	p.quitting = true
	log.Printf("quit: %s is quitting.", p.Name())
	return true
}

func (p *Player) sneeze(cmd *command.Command) (handled bool) {
	cmd.Respond("You sneeze. Aaahhhccchhhooo!")
	cmd.Broadcast([]thing.Interface{p}, "You see %s sneeze.", cmd.Issuer.Name())
	PlayerList.Broadcast(p.Locate().List(), "You hear a loud sneeze.")
	return true
}

func (p *Player) who(cmd *command.Command) (handled bool) {
	cmd.Broadcast([]thing.Interface{p}, "You see %s concentrate.", p.Name())
	msg := ""

	c := 1
	for _, p := range PlayerList.List(p) {
		msg += fmt.Sprintf("  %s\n", p.Name())
		c++
	}
	msg += fmt.Sprintf(" TOTAL PLAYERS: %d", c)

	if c < 2 {
		msg = "You are all alone in this world."
	}

	cmd.Respond(msg)
	return true
}
