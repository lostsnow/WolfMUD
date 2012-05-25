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
	"wolfmud.org/utils/UID"
	"wolfmud.org/utils/broadcaster"
	"wolfmud.org/utils/command"
	"wolfmud.org/utils/sender"
)

var (
	playerCount = 0
)

type Interface interface {
}

type Player struct {
	*mobile.Mobile
	sender sender.Interface
	world  broadcaster.Interface
	id     UID.UID
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
	p.id = p.Mobile.Thing.UniqueId()

	// Put player into the world, announce and describe location
	world.AddThing(p)
	p.Locate().Broadcast([]thing.Interface{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.Name())
	p.Parse("LOOK")

	PlayerList.Add(p)

	runtime.SetFinalizer(p, Final)

	return p
}

func Final(p *Player) {
	log.Printf("+++ Player %d finalized +++\n", p.id)
}

func (p *Player) Destroy() {

	name := p.Name()

	log.Printf("Destroying player: %s\n", name)

	for !p.remove() {
	}

	p.world.Broadcast(nil, "AAAaaarrrggghhh!!!\nA scream is heard across the land as %s is unceremoniously extracted from the world.", name)

	p.world = nil
	p.sender = nil
	p.Mobile = nil

	log.Printf("Destroyed player: %s\n", name)
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

// Parse parses commands passed to delegates handling of the command. To
// avoid deadlocks, inconsistencies, races and other unmentionables we lock
// the location of the player. There is a race condition between getting the
// player's location and locking it - they may have moved in-between. We
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
	cmd.Relock = p.Locate()

	for retry := false; cmd.Relock != nil || retry; {
		cmd.AddLock()
		retry = p.subParse(cmd)
	}

}

func (p *Player) subParse(cmd *command.Command) (retry bool) {
	for _, l := range cmd.Locks {
		if t, ok := l.(location.Interface); ok {
			t.Lock()
			defer func() {
				t.Unlock()
			}()
		}
	}
	if cmd.IsLocked(p.Locate()) {
		handled := p.Process(cmd)
		if handled == false && cmd.Relock == nil {
			p.Respond("Eh?")
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
		fmt.Printf("player.Respond: Player %d is a Zombie\n", p.id)
	}
}

func (p *Player) Process(cmd *command.Command) (handled bool) {

	switch cmd.Verb {
	default:
		handled = p.Mobile.Process(cmd)
	case "SNEEZE":
		handled = p.sneeze(cmd)
	case "MEMPROF":
		handled = p.memprof(cmd)
	case "WHO":
		handled = p.who(cmd)
	}

	return
}

func (p *Player) sneeze(cmd *command.Command) (handled bool) {
	p.Respond("You sneeze. Aaahhhccchhhooo!")
	p.world.Broadcast([]thing.Interface{p}, "You hear a loud sneeze.")
	return true
}

func (p *Player) memprof(cmd *command.Command) (handled bool) {
	f, err := os.Create("memprof")
	if err != nil {
		p.Respond("Memory Profile Not Dumped: %s", err)
		return false
	}
	pprof.WriteHeapProfile(f)
	f.Close()

	cmd.Respond("Memory profile dumped")
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

	p.Respond(msg)
	return true
}
