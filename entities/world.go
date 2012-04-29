package entities

import (
	"fmt"
	"net"
	"strconv"
	"runtime"
)

type stats struct {
	Alloc       uint64
	HeapObjects uint64
	Goroutines  int
}

var (
	orig *stats
	old  *stats
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
	players   map[string]Player
}

func NewWorld() World {
	return &world{
		players: make(map[string]Player, 10),
	}
}

func (w *world) Start() {

	fmt.Println("Starting WolfMUD server...")

	ln, err := net.Listen("tcp", "localhost:4001")
	if err != nil {
		fmt.Printf("world.Start: Error setting up listener, %s\nServer will now exit.\n", err)
		return
	}

	fmt.Println("Accepting connections.")
	w.Stats()

	for {
		if conn, err := ln.Accept(); err != nil {
			fmt.Printf("world.Start: Error accepting connection: %s\nServer will now exit.\n", err)
			return
		} else {
			fmt.Printf("world.Start: connection from %s.\n", conn.RemoteAddr().String())
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
		conn,
	)

	w.players[p.Alias()] = p
	w.locations[0].Add(p)

	fmt.Printf("world.AddPlayer: connection %s allocated %s, %d players online.\n", conn.RemoteAddr().String(), p.Name(), len(w.players))

	go p.Start()

	w.Stats()
}

func (w *world) RemovePlayer(alias string) {
	p := w.players[alias]
	p.Where().Remove(alias, 1)
	delete(w.players, alias)
	fmt.Printf("world.RemovePlayer: removing %s, %d players online.\n", alias, len(w.players))
	w.Stats()
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

func (w *world) Stats() {
	runtime.GC()
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)
	ng := runtime.NumGoroutine()

	if old == nil {
		old = new(stats)
		old.Alloc = m.Alloc
		old.HeapObjects = m.HeapObjects
		old.Goroutines = ng
	}

	if orig == nil {
		orig = new(stats)
		orig.Alloc = m.Alloc
		orig.HeapObjects = m.HeapObjects
		orig.Goroutines = ng
	}

	fmt.Printf("Alloc: %d (%d/%d), HeapObjects: %d (%d/%d), Go Routines: %d (%d/%d)\n", m.Alloc, int(m.Alloc-old.Alloc), int(m.Alloc-orig.Alloc), m.HeapObjects, int(m.HeapObjects-old.HeapObjects), int(m.HeapObjects-orig.HeapObjects), ng, ng-old.Goroutines, ng-orig.Goroutines)

	old.Alloc = m.Alloc
	old.HeapObjects = m.HeapObjects
	old.Goroutines = ng
}
