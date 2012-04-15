package player

import (
	"fmt"
	"strings"
	"wolfmud.org/location"
)

type Player struct {
	name          string
	descrition    string
	location      location.Location
	Send          chan string
	terminalWidth int
}

func New(n string, l location.Location) (p Player) {
	defer func() { go p.send(p.Send) }()

	return Player{
		name:          n,
		descrition:    "An adventurer like yourself",
		location:      l,
		Send:          make(chan string, 10),
		terminalWidth: 80,
	}
}

func (p *Player) send(c <-chan string) {
	for msg := range c {
		for msg != "" {
			if len(msg) <= p.terminalWidth {
				fmt.Println(msg)
				msg = ""
			} else {
				fmt.Println(msg[0:p.terminalWidth])
				msg = msg[p.terminalWidth:]
			}
		}
	}
}

func (p *Player) Parse(cmd string) {
	fmt.Printf("> %s\n", cmd)
	tok := strings.Split(cmd, ` `)
	cmd = strings.ToUpper(tok[0])
	args := tok[1:]

	// See if current location can handle command
	if handled := p.location.Command(cmd, args); handled == true {
		return
	}

	// See if player can handle command
	if handled := p.Command(cmd, args); handled == true {
		return
	}

	// Unknown command
	p.Send <- fmt.Sprintf("You don't know how to '%s'\n", cmd)
}

func (p *Player) Command(cmd string, args []string) (handled bool) {
	switch cmd {
	case `LOOK`:
		handled = p.Look(args)
	case `SAY`:
		handled = p.Say(args)
	}
	return
}

func (p *Player) Look(args []string) (handled bool) {
	if len(args) == 0 {
		return
	}

	p.Send <- fmt.Sprintf("You look at %s\n", p.name)
	p.Send <- fmt.Sprintf("%s\n", p.descrition)
	return true
}

func (p *Player) Say(args []string) (handled bool) {
	if len(args) == 0 {
		p.Send <- fmt.Sprintf("You go to say something but can't remember what it was...\n")
	} else {
		p.Send <- fmt.Sprintf("You say \"%s\"\n", strings.Join(args, ` `))
	}
	return true
}
