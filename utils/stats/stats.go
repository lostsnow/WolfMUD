package stats

import (
	"fmt"
	"runtime"
	"time"
	"wolfmud.org/entities/mobile/player"
)

type stats struct {
	Alloc       uint64
	HeapObjects uint64
	Goroutines  int
	MaxPlayers  int
}

type Stats struct {
	orig *stats
	old  *stats
}

func Start() {
	c := time.Tick(5 * time.Second)
	s := new(Stats)
	go func() {
		for _ = range c {
			s.stats()
		}
	}()

	// 1st time initialisation
	s.stats()
}

func (s *Stats) stats() {
	runtime.GC()
	runtime.Gosched()

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	ng := runtime.NumGoroutine()

	pl := player.PlayerList.Length()

	if s.old == nil {
		s.old = new(stats)
		s.old.Alloc = m.Alloc
		s.old.HeapObjects = m.HeapObjects
		s.old.Goroutines = ng
		s.old.MaxPlayers = pl
	}

	if s.orig == nil {
		s.orig = new(stats)
		s.orig.Alloc = m.Alloc
		s.orig.HeapObjects = m.HeapObjects
		s.orig.Goroutines = ng
		s.orig.MaxPlayers = pl
	}

	if s.old.MaxPlayers < pl {
		s.old.MaxPlayers = pl
	}

	fmt.Printf("%s: %12d A[%+9d %+9d] %12d HO[%+6d %+6d] %6d GO[%+6d %+6d] %4d PL[%4d]\n", time.Now().Format(time.Stamp), m.Alloc, int(m.Alloc-s.old.Alloc), int(m.Alloc-s.orig.Alloc), m.HeapObjects, int(m.HeapObjects-s.old.HeapObjects), int(m.HeapObjects-s.orig.HeapObjects), ng, ng-s.old.Goroutines, ng-s.orig.Goroutines, pl, s.old.MaxPlayers)

	s.old.Alloc = m.Alloc
	s.old.HeapObjects = m.HeapObjects
	s.old.Goroutines = ng
}
