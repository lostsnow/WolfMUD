package entities

import (
	"fmt"
	"net"
	"strings"
)

type Player interface {
	Mobile
	Responder
	Start()
	Output()
}

type player struct {
	mobile
	responder
	world    World
	output   chan string
	quit     chan bool
	quitting bool
	conn     net.Conn
}

func NewPlayer(w World, name, alias, description string, conn net.Conn) Player {
	return &player{
		mobile:   *NewMobile(name, alias, description).(*mobile),
		world:    w,
		output:   make(chan string, 10),
		quit:     make(chan bool),
		quitting: false,
		conn:     conn,
	}
}

func (p *player) Destroy() {

	name := p.name

	p.quit <- true
	close(p.quit)

	fmt.Printf("Closing socket for %s\n", name)
	if err := p.conn.Close(); err != nil {
		fmt.Printf("Error closing socket for %s, %s\n", name, err)
	}

	p.world.RemovePlayer(p.alias)
	p.world = nil
	//p.mobile = nil

	//	for drain := range p.output {
	//		_ = drain
	//	}
	close(p.output)

	fmt.Printf("Destroyed player: %s\n", name)
}

func (p *player) Start() {

	var inBuffer [255]byte

	// Start async output handler
	go p.Output()

	p.conn.Write([]byte("\n\nWelcome To WolfMUD\n\n"))
	p.Where().RespondGroup([]Thing{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.name)
	p.Process(NewCommand(p, "LOOK"))

	for {
		if b, err := p.conn.Read(inBuffer[0:254]); err != nil {
			p.quitting = true
			fmt.Printf("player.Start: Comms error for: %s, %s\n", p.name, err)
			if l := p.location; l != nil {
				p.world.RespondGroup([]Thing{p}, "\nAAAaaarrrggghhh!!!\nA scream is heard across the land as %s is unceremoniously extracted from the world.", p.alias)
			}
			p.Destroy()
			return
		} else {
			input := strings.TrimSpace(string(inBuffer[0:b]))
			cmd := NewCommand(p, input)
			if handled := p.Process(cmd); handled == false {
				p.Respond("Eh?")
			}
		}
	}

}

func (p *player) Output() {
	for {
		select {
		case <-p.quit:
			fmt.Printf("player.Output: handler ending for %s\n", p.name)
			return
		case s := <-p.output:
			if p.quitting == false {
				if _, err := p.conn.Write([]byte(s)); err != nil {
					fmt.Printf("player.Output: Comms error for: %s, %s\n", p.name, err)
				}
			}
		}
	}
}

func (p *player) Respond(format string, any ...interface{}) {
	if p.quitting == false {
		p.output <- fmt.Sprintf("\n"+format+"\n> ", any...)
	}
	return
}

func (p *player) RespondGroup(ommit []Thing, format string, any ...interface{}) {
	p.location.RespondGroup(ommit, format, any...)
	return
}
