package entities

import (
	"fmt"
	"net"
	"runtime"
	"strings"
)

type Player interface {
	Mobile
	Responder
	Start(conn net.Conn)
	Output(conn net.Conn, quit <-chan bool)
}

type player struct {
	mobile
	responder
	world  World
	output chan string
	conn   net.Conn
}

func NewPlayer(w World, name, alias, description string) Player {
	return &player{
		mobile: *NewMobile(name, alias, description).(*mobile),
		world:  w,
		output: make(chan string, 10),
	}
}

func (p *player) Start(conn net.Conn) {

	var inBuffer [255]byte

	// Start async output handler with quit channel
	quit := make(chan bool)
	go p.Output(conn, quit)

	conn.Write([]byte("\n\nWelcome To WolfMUD\n\n"))
	p.Where().RespondGroup([]Thing{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.name)
	p.Process(NewCommand(p, "LOOK"))

	for {
		if b, err := conn.Read(inBuffer[0:254]); err != nil {
			fmt.Printf("player.Start: Comms error for: %s, %s\n", p.name, err)
			if l := p.location; l != nil {
				l.Remove(p.alias, 1)
				p.world.RespondGroup([]Thing{p}, "\nAAAaaarrrggghhh!!!\nA scream is heard across the land as %s is unceremoniously extracted from the world.", p.name)
			}
			fmt.Printf("Releasing player: %s\n", p.name)
			break
		} else {
			input := strings.TrimSpace(string(inBuffer[0:b]))
			cmd := NewCommand(p, input)
			if handled := p.Process(cmd); handled == false {
				p.Respond("Eh?")
			}
		}
	}

	quit <- true
	close(quit)

	for drain := range p.output {
		_ = drain
	}
	close(p.output)

	fmt.Printf("Closing socket for %s\n", p.alias)
	if err := conn.Close(); err != nil {
		fmt.Printf("Error closing socket for %s, %s\n", p.name, err)
	}

	runtime.GC()

}

func (p *player) Output(conn net.Conn, quit <-chan bool) {
	for {
		select {
		case <-quit:
			fmt.Printf("Output handler ending for %s\n", p.name)
			return
		case s := <-p.output:
			if _, err := conn.Write([]byte(s)); err != nil {
				fmt.Printf("player.Output: Comms error for: %s, %s\n", p.name, err)
			}
		}
	}
}

func (p *player) Respond(format string, any ...interface{}) {
	p.output <- fmt.Sprintf("\n"+format+"\n> ", any...)
	return
}

func (p *player) RespondGroup(ommit []Thing, format string, any ...interface{}) {
	p.location.RespondGroup(ommit, format, any...)
	return
}
