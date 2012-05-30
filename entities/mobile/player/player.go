package player

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
	"wolfmud.org/entities/mobile"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/broadcaster"
	"wolfmud.org/utils/command"
	"wolfmud.org/utils/sender"
)

var (
	playerCount = 0
)

type Player struct {
	*mobile.Mobile
	sender   sender.Interface
	world    broadcaster.Interface
	name     string
	quitting bool
}

func New(sender sender.Interface, world broadcaster.Interface) *Player {

	playerCount++
	postfix := strconv.Itoa(playerCount)

	name := "Player " + postfix
	alias := []string{"PLAYER" + postfix}
	description := "This is Player " + postfix + "."

	p := &Player{
		Mobile: mobile.New(name, alias, description),
		sender: sender,
		world:  world,
	}
	p.name = p.Name()

	// Put player into the world, announce and describe location
	world.AddThing(p)
	p.Locate().Broadcast([]thing.Interface{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.Name())
	p.Parse("LOOK")

	PlayerList.Add(p)

	log.Printf("Player %d created: %s\n", p.UniqueId(), p.Name())
	runtime.SetFinalizer(p, final)

	return p
}

func final(p *Player) {
	log.Printf("+++ %s finalized +++\n", p.name)
}

func (p *Player) Quitting() bool {
	p.Lock()
	defer p.Unlock()
	return p.quitting
}

func (p *Player) Destroy() {

	name := p.Name()

	log.Printf("Destroy: %s\n", name)

	if p.Quitting() {
		log.Printf("%s is quitting @ %s", name, p.Locate().Name())
		p.Locate().Broadcast(nil, "%s gives a strangled cry of 'Bye Bye', and then slowly fades away and is gone.", name)
	}

	for !p.remove() {
	}

	if !p.Quitting() {
		p.world.Broadcast(nil, "AAAaaarrrggghhh!!!\nA scream is heard across the land as %s is unceremoniously extracted from the world.", name)
	}

	p.world = nil
	p.sender = nil
	p.Mobile = nil

	log.Printf("Destroyed: %s\n", name)
}

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

// Parse takes a string and begins the delegation to potential processors. To
// avoid deadlocks, inconsistencies, race conditions and other unmentionables we
// lock the location of the player. However there is a race condition between getting
// the player's location and locking it - they may have moved in-between. We
// therefore get and lock their current location then check it's still their
// current location. If it is not we unlock and try again.
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
func (p *Player) Parse(input string) {

	cmd := command.New(p, input)
	cmd.AddLock(p.Locate())

	for retry := cmd.LocksModified(); retry; {
		retry = p.parseStage2(cmd)
	}

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
	p.quitting = true
	log.Printf("quit: %s is quitting.", p.Name())
	return true
}

func (p *Player) sneeze(cmd *command.Command) (handled bool) {
	cmd.Respond("You sneeze. Aaahhhccchhhooo!")
	p.Locate().Broadcast([]thing.Interface{p}, "You see %s sneeze.", cmd.Issuer.Name())
	p.world.Broadcast(p.Locate().List(), "You hear a loud sneeze.")
	return true
}

func (p *Player) who(cmd *command.Command) (handled bool) {
	p.Locate().Broadcast([]thing.Interface{p}, "You see %s concentrate.", p.Name())
	msg := ""

	for _, p := range PlayerList.List(p) {
		msg += fmt.Sprintf("  %s\n", p.Name())
	}

	if len(msg) == 0 {
		msg = "You are all alone in this world."
	}

	cmd.Respond(msg)
	return true
}
