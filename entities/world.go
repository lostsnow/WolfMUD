package entities

import (
	"fmt"
	"net"
	"runtime"
	"time"
	//"os"
	//"runtime/pprof"
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

type World interface {
	Responder
	Start()
	AddPlayer(player Player)
	RemovePlayer(player Player)
	AddLocation(l Location)
}

type world struct {
	locations   []Location
	players     []Player
	playersLock chan bool
}

func NewWorld() World {
	return &world{
		playersLock: make(chan bool, 1),
	}
}

func (w *world) Start() {

	fmt.Println("Starting WolfMUD server...")

	ta, err := net.ResolveTCPAddr("tcp", "localhost:4001");
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

	// Setup stat ticker
	c := time.Tick(5 * time.Second)
	go func() {
		for _ = range c {
			w.Stats()
		}
	}()

	w.Stats()

	//cx := 0

	for {
		if conn, err := ln.AcceptTCP(); err != nil {
			fmt.Printf("world.Start: Error accepting connection: %s\nServer will now exit.\n", err)
			return
		} else {
			fmt.Printf("world.Start: connection from %s.\n", conn.RemoteAddr().String())
			w.startPlayer(conn)

			/*
			cx++
			if cx == 1000 {
				f, _ := os.Create("memprofile")
				pprof.WriteHeapProfile(f)
				f.Close()
			}
			*/

		}
	}
}

func (w *world) startPlayer(conn *net.TCPConn) {
	c := NewClient(conn)
	p := NewPlayer(w)

	p.AttachClient(c)

	c.SendPlain("\n\nWelcome to WolfMUD\n\n")
	p.Parse("LOOK")
	p.Where().RespondGroup([]Thing{p}, "There is a puff of smoke and %s appears spluttering and coughing.", p.Name())

	fmt.Printf("world.AddPlayer: connection %s allocated %s, %d players online.\n", conn.RemoteAddr().String(), p.Name(), len(w.players))

	go c.Start()
}

func (w *world) AddPlayer(p Player) {
	w.playersLock <- true
	defer func() {
		<-w.playersLock
	}()

	w.players = append(w.players, p)
	w.locations[0].Add(p)
}

func (w *world) RemovePlayer(player Player) {
	name := player.Name()
	w.playersLock <- true
	defer func() {
		<-w.playersLock
	}()

	for i, p := range w.players {
		if p == player {
			if l := p.Where(); l == nil {
				fmt.Printf("world.RemovePlayer: Eeep! %s is nowhere!.\n", name)
			} else {
				l.Remove(player.Alias(), 1)
			}
			w.players = append(w.players[0:i], w.players[i+1:]...)
			fmt.Printf("world.RemovePlayer: removing %s, %d players online.\n", name, len(w.players))
			return
		}
	}
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
	runtime.Gosched()
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

	fmt.Printf("%s: %12d A[%+9d %+9d] %12d HO[%+6d %+6d] %6d GO[%+6d %+6d]  %6d PL\n", time.Now().Format(time.Stamp), m.Alloc, int(m.Alloc-old.Alloc), int(m.Alloc-orig.Alloc), m.HeapObjects, int(m.HeapObjects-old.HeapObjects), int(m.HeapObjects-orig.HeapObjects), ng, ng-old.Goroutines, ng-orig.Goroutines, len(w.players))

	old.Alloc = m.Alloc
	old.HeapObjects = m.HeapObjects
	old.Goroutines = ng
}
