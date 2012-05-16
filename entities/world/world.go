package world

import (
	"fmt"
	"net"
	"wolfmud.org/client"
	"wolfmud.org/entities/location"
	"wolfmud.org/entities/mobile/player"
	"wolfmud.org/entities/thing"
	"wolfmud.org/utils/broadcaster"
	"wolfmud.org/utils/stats"
)

type Interface interface {
	broadcaster.Interface
	AddLocation(l location.Interface)
	Start()
}

type World struct {
	locations []location.Interface
}

func New() *World {
	return &World{}
}

func (w *World) Start() {

	fmt.Println("Starting WolfMUD server...")

	ta, err := net.ResolveTCPAddr("tcp", "localhost:4001")
	if err != nil {
		fmt.Printf("world.Start: Error resolving TCP address, %s\nServer will now exit.\n", err)
		return
	}

	ln, err := net.ListenTCP("tcp", ta)
	if err != nil {
		fmt.Printf("world.Start: Error setting up listener, %s\nServer will now exit.\n", err)
		return
	}

	fmt.Println("Accepting connections.")

	stats.Start()

	for {
		if conn, err := ln.AcceptTCP(); err != nil {
			fmt.Printf("world.Start: Error accepting connection: %s\nServer will now exit.\n", err)
			return
		} else {
			fmt.Printf("world.Start: connection from %s.\n", conn.RemoteAddr().String())
			w.startPlayer(conn)
		}
	}
}

func (w *World) startPlayer(conn *net.TCPConn) {
	c := client.New(conn)
	p := player.New(w)

	p.AttachClient(c)

	c.SendWithoutPrompt(`

WolfMUD © 2012 Andrew 'Diddymus' Rolfe

    World
    Of
    Living
    Fantasy

`)
	w.locations[0].Add(p)
	p.Parse("LOOK")
	w.locations[0].Broadcast([]thing.Interface{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.Name())

	fmt.Printf("world.startPlayer: connection %s allocated %s, %d players online.\n", conn.RemoteAddr().String(), p.Name(), player.PlayerList.Length())

	go c.Start()
}

func (w *World) AddLocation(l location.Interface) {
	w.locations = append(w.locations, l)
}

func (w *World) Broadcast(ommit []thing.Interface, format string, any ...interface{}) {
	fmt.Println("World broadcast: %#v", player.PlayerList.List())

	msg := fmt.Sprintf("\n"+format, any...)

OMMIT:
	for _, p := range player.PlayerList.List() {
		fmt.Printf("Checking: %s\n", p.Name())
		for _, o := range ommit {
			if o.IsAlso(p) {
				fmt.Printf("Ommiting: %s\n", p.Name())
				continue OMMIT
			}
		}
		p.Respond(msg)
	}
}
