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
	"wolfmud.org/utils/UID"
//	"wolfmud.org/utils/broadcaster"
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
	//world  broadcaster.Interface
	lock   chan bool
	id     UID.UID
}

func New(sender sender.Interface) *Player {

	playerCount++
	postfix := strconv.Itoa(playerCount)

	name := "Player " + postfix
	alias := []string{"PLAYER" + postfix}
	description := "This is Player " + postfix + "."

	p := &Player{
		Mobile: mobile.New(name, alias, description),
		sender: sender,
		//world:  world,
		//lock:   make(chan bool, 1),
	}
	p.id = p.Mobile.Thing.UniqueId()

	//PlayerList.Add(p)

	runtime.SetFinalizer(p, Final)

	return p
}

func Final(p *Player) {
	log.Printf("+++ Player %d finalized +++\n", p.id)
}

func (p *Player) destroy() {

	name := p.Name()

	log.Printf("Destroying player: %s\n", name)

	//p.Locate().Remove(p)
	//PlayerList.Remove(p)

	//p.world.Broadcast(nil, "AAAaaarrrggghhh!!!\nA scream is heard across the land as %s is unceremoniously extracted from the world.", name)

	//p.world = nil
	p.sender = nil
	p.Mobile = nil

	log.Printf("Destroyed player: %s\n", name)
}

func (p *Player) Parse(input string) {
	handled := p.Process(command.New(p, input))
	if handled == false {
		p.Respond("Eh? %s?", input)
	}
}

func (p *Player) Respond(format string, any ...interface{}) {
	if c := p.sender; c != nil {
		c.Send(format, any...)
		runtime.Gosched()
	} else {
		fmt.Printf("player.Respond: %s is a Zombie\n", p.Name())
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
	//p.world.Broadcast([]thing.Interface{p}, "You hear a loud sneeze.")
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
