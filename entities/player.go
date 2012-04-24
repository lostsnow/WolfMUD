package entities

import (
	"fmt"
	"net"
	"strings"
)

type Player interface {
	Mobile
	Responder
	Run(conn net.Conn)
	Input(text string)
	Output(conn net.Conn)
}

type player struct {
	mobile
	responder
	input  chan string
	output chan string
	conn   net.Conn
}

func NewPlayer(name, alias, description string) Player {

	p := &player{
		mobile: *NewMobile(name, alias, description).(*mobile),
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

func (p *player) Run(conn net.Conn) {

	go p.Output(conn)

	conn.Write([]byte("\n\nWelcome To WolfMUD\n\n"))
	p.Where().RespondGroup([]Thing{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.Name())
	p.Input("LOOK")

	for {
		var buffer [255]byte
		b, _ := conn.Read(buffer[0:254])
		p.Input(string(buffer[0:b]))
	}
	conn.Close()
}

func (p *player) Input(text string) {
	text = strings.TrimSpace(text)
	p.input <- text
	println("Received [" + text + "]")
}

func (p *player) Output(conn net.Conn) {
	for {
		select {
		case s := <-p.output:
			conn.Write([]byte(s))
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
