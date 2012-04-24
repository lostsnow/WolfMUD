package entities

import (
	"fmt"
)

type Player interface {
	Mobile
	Responder
	Input(text string)
	Output() (text string)
}

type player struct {
	mobile
	responder
	input  chan string
	output chan string
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

func (p *player) Input(text string) {
	p.input <- text
}

func (p *player) Output() (text string) {
	select {
	default:
	case s := <-p.output:
		text += s
	}
	return
}

func (p *player) Respond(format string, any ...interface{}) {
	p.output <- fmt.Sprintf(format, any...)
	return
}
