package entities

import (
	"fmt"
	"net"
	"strconv"
)

var playerCount = 0

type World interface {
	Responder
	Start()
	AddPlayer(conn net.Conn)
	RemovePlayer(alias string)
	AddLocation(l Location)
}

type world struct {
	locations []Location
	players   []Player
}

func NewWorld() World {
	return &world{}
}

func (w *world) Start() {

	fmt.Println("Starting WolfMUD server...")

	ln, err := net.Listen("tcp", "localhost:4001")
	if err != nil {
		fmt.Printf("server.main: Error setting up listener, %s\nServer will now exit.\n", err)
		return
	}

	fmt.Println("Accepting connections.")

	for {
		if conn, err := ln.Accept(); err != nil {
			fmt.Printf("server.main: Error accepting connection: %s\nServer will now exit.\n", err)
			return
		} else {
			w.AddPlayer(conn)
		}
	}
}

func (w *world) AddPlayer(conn net.Conn) {
	playerCount++
	postfix := strconv.Itoa(playerCount)

	p := NewPlayer(
		w,
		"Player "+postfix,
		"PLAYER"+postfix,
		"This is Player "+postfix+".",
	)

	fmt.Printf("Connection from: %s, allocated %s\n", conn.RemoteAddr().String(), p.Name())

	w.players = append(w.players, p)
	w.locations[0].Add(p)
	go p.Start(conn)
}

func (w *world) RemovePlayer(alias string) {
}

func (w *world) AddLocation(l Location) {
	w.locations = append(w.locations, l)
}

func (w *world) Respond(format string, any ...interface{}) {
	msg := fmt.Sprintf(format, any...)
	for _, p := range w.players {
		p.Respond(msg)
	}
}

func (w *world) RespondGroup(ommit []Thing, format string, any ...interface{}) {
	msg := fmt.Sprintf(format, any...)

OMMIT:
	for _, p := range w.players {
		for _, o := range ommit {
			if o.IsAlso(p) {
				continue OMMIT
			}
			p.Respond(msg)
		}
	}
}
