package entities

import (
	"fmt"
	"strconv"
)

var (
	playerCount = 0
)

type Player interface {
	Mobile
	Responder
	AttachClient(client Client)
	DetachClient()
	Destroy()
}

type player struct {
	mobile
	responder
	world  World
	client Client
	lock   chan bool
}

func NewPlayer(w World) Player {

	playerCount++
	postfix := strconv.Itoa(playerCount)

	name := "Player " + postfix
	alias := "PLAYER" + postfix
	description := "This is Player " + postfix + "."

	p := &player{
		mobile: *NewMobile(name, alias, description).(*mobile),
		world:  w,
		lock:   make(chan bool, 1),
	}

	w.AddPlayer(p)

	return p
}

func (p *player) AttachClient(client Client) {
	p.lock <- true
	defer func(){
		<-p.lock
	}()
	p.client = client
	client.AttachPlayer(p)
}

func (p *player) DetachClient() {
	p.lock <- true
	defer func(){
		<-p.lock
	}()
	p.client = nil
}

func (p *player) hasClient() bool {
	p.lock <- true
	defer func(){
		<-p.lock
	}()
	return (p.client != nil)
}

func (p *player) Destroy() {

	name := p.name
	world := p.world

	fmt.Printf("Destroying player: %s\n", name)

	p.world.RemovePlayer(p)
	p.world = nil
	p.DetachClient()

	world.RespondGroup(nil, "\nAAAaaarrrggghhh!!!\nA scream is heard across the land as %s is unceremoniously extracted from the world.", name)

	fmt.Printf("Destroyed player: %s\n", name)
}

func (p *player) Parse(input string) {
	handled := p.Process(NewCommand(p, input))
	if handled == false {
		p.Respond("Eh? %s?", input)
	}
}

func (p *player) Respond(format string, any ...interface{}) {
	if c := p.client; c != nil {
		c.SendResponse(format, any...)
	} else {
		fmt.Printf("player.Respond: %s is a Zombie\n", p.name)
	}
}

func (p *player) RespondGroup(ommit []Thing, format string, any ...interface{}) {
	p.location.RespondGroup(ommit, format, any...)
	return
}
