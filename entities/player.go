package entities

import (
	"fmt"
	"net"
	"strings"
)

type Player interface {
	Mobile
	Responder
	Start(conn net.Conn)
	Input(text string)
	Output(conn net.Conn)
}

type player struct {
	mobile
	responder
	world  World
	input  chan string
	output chan string
	conn   net.Conn
}

func NewPlayer(w World, name, alias, description string) Player {

	p := &player{
		mobile: *NewMobile(name, alias, description).(*mobile),
		world:  w,
		input:  make(chan string, 10),
		output: make(chan string, 10),
	}

	defer func() {
		go func() {
			for {
				if handled := p.Process(NewCommand(p, <-p.input)); handled == false {
					p.Respond("Eh?")
				}
			}
		}()
	}()

	return p
}

func (p *player) Start(conn net.Conn) {

	go p.Output(conn)

	var buffer [255]byte

	conn.Write([]byte("\n\nWelcome To WolfMUD\n\n"))
	p.Where().RespondGroup([]Thing{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.Name())
	p.Input("LOOK")

	for {
		if b, err := conn.Read(buffer[0:254]); err != nil {
			fmt.Printf("player.Start: Comms error for: %s, %s\n", p.Name(), err)
			if l := p.location; l != nil {
				l.Remove(p.Alias(), 1)
				p.world.RespondGroup([]Thing{p}, "\nAAAaaarrrggghhh!!!\nA scream is heard across the land as %s is unceremoniously extracted from the world.", p.Alias())
			}
			fmt.Printf("Releasing player: %s\n", p.Name())
			conn.Close()
			return
		} else {
			p.Input(string(buffer[0:b]))
		}
	}
}

func (p *player) Input(text string) {
	text = strings.TrimSpace(text)
	p.input <- text
}

func (p *player) Output(conn net.Conn) {
	for {
		select {
		case s := <-p.output:
			if _, err := conn.Write([]byte(s)); err != nil {
				fmt.Printf("player.Output: Comms error for: %s, %s\n", p.Name(), err)
				return
			}
		}
	}
}

func (p *player) Respond(format string, any ...interface{}) {
	p.output <- fmt.Sprintf(format+"\n> ", any...)
	return
}

func (p *player) RespondGroup(ommit []Thing, format string, any ...interface{}) {
	p.location.RespondGroup(ommit, format, any...)
	return
}
