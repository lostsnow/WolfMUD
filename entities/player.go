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
}

type player struct {
	mobile
	responder
	world    World
	conn     net.Conn
	connLock chan bool
}

func NewPlayer(w World, name, alias, description string, conn net.Conn) Player {
	return &player{
		mobile:   *NewMobile(name, alias, description).(*mobile),
		world:    w,
		conn:     conn,
		connLock: make(chan bool, 1),
	}
}

func (p *player) connError() (cerr bool) {

	// Lock and make sure we unlock
	p.connLock <- true
	defer func() {
		<-p.connLock
	}()

	cerr = (p.conn == nil)
	return
}

func (p *player) setConnError(err error) {

	// Lock and make sure we unlock
	p.connLock <- true
	defer func() {
		<-p.connLock
	}()

	// Make sure error has not already been flagged
	if p.conn == nil {
		return
	}

	fmt.Printf("player.setConnError: Comms error for: %s, %s\n", p.name, err)
	p.world.RespondGroup([]Thing{p}, "\nAAAaaarrrggghhh!!!\nA scream is heard across the land as %s is unceremoniously extracted from the world.", p.alias)
	p.world.RemovePlayer(p.alias)

	fmt.Printf("player.setConnError: Closing socket for %s\n", p.name)
	if err := p.conn.Close(); err != nil {
		fmt.Printf("player.setConnError: Error closing socket for %s, %s\n", p.name, err)
	}
	p.conn = nil
}

func (p *player) Destroy() {

	name := p.name

	p.world.RemovePlayer(p.alias)
	p.world = nil

	fmt.Printf("Destroyed player: %s\n", name)
}

func (p *player) Start() {

	var inBuffer [255]byte

	p.conn.Write([]byte("\n\nWelcome To WolfMUD\n\n"))
	p.Where().RespondGroup([]Thing{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.name)
	p.Process(NewCommand(p, "LOOK"))

	for {
		if p.connError() {
			return
		}
		if b, err := p.conn.Read(inBuffer[0:254]); err != nil {
			p.setConnError(err)
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

func (p *player) Respond(format string, any ...interface{}) {
	if p.connError() {
		return
	}
	s := fmt.Sprintf("\n"+format+"\n>", any...)
	if _, err := p.conn.Write([]byte(s)); err != nil {
		p.setConnError(err)
		fmt.Printf("player.asyncOutput: Comms error for: %s, %s\n", p.name, err)
	}
	return
}

func (p *player) RespondGroup(ommit []Thing, format string, any ...interface{}) {
	p.location.RespondGroup(ommit, format, any...)
	return
}
