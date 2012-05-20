// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights resolved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package stats implements periodic collection and display of various -
// possibly interisting - statistics. A typical reading might be:
//
//	2012/05/18 14:32:19       448024 A[   +40624    +45416]          483 HO[   +50    +64]      8 GO[    +2     +2]    1 PL[   1]
//
// This shows:
//
//	448024 A[   +40624    +45416] - allocated memory, change since last collection, change since start
//	        483 HO[   +50    +64] - heap objects, change since last collection, change since start
//	          8 GO[    +2     +2] - Goroutines, change since last collection, change since start
//	                   1 PL[   1] - Number of players, maximum players logged on at once
package stats

import (
	"log"
	"runtime"
	"time"
	"wolfmud.org/entities/mobile/player"
)

// Interval can be changed before calling Start to set a different collection
// interval.
//
// TODO: When we have sorted out global settings this needs moving there.
var (
	Interval = 10 * time.Second // Time  between collections
)

// record represents a single collection of statistical data.
type record struct {
	Alloc       uint64
	HeapObjects uint64
	Goroutines  int
	MaxPlayers  int
}

// save records a set of data into a record type.
func (s *record) save(alloc, heap uint64, goroutines, maxPlayers int) {
	s.Alloc = alloc
	s.HeapObjects = heap
	s.Goroutines = goroutines
	s.MaxPlayers = maxPlayers
}

// state is used to hold record data between each collection run
type state struct {
	s *record // Original starting stats
	o *record // Old stats from previous loop
}

// Start begins collection and reporting of statistics. The interval between
// updates is controlled via the stats.Interval variable.
func Start() {
	c := time.Tick(Interval)
	s := &state{&record{}, &record{}}

	go func() {
		for _ = range c {
			s.collect()
		}
	}()

	s.collect() // 1st time initialisation
}

// collect runs periodically to collect, process and report statistics.
func (s *state) collect() {
	runtime.GC()
	runtime.Gosched()

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	ng := runtime.NumGoroutine()
	pl := player.PlayerList.Length()

	if s.s.Alloc == 0 {
		s.s.save(m.Alloc, m.HeapObjects, ng, pl)
		s.o.save(m.Alloc, m.HeapObjects, ng, pl)
	}

	// Calculate difference in resources since last run
	ad := int64(m.Alloc - s.o.Alloc)
	hd := int(m.HeapObjects - s.o.HeapObjects)
	gd := ng - s.o.Goroutines

	// Calculate difference in resources since start
	as := int64(m.Alloc - s.s.Alloc)
	hs := int(m.HeapObjects - s.s.HeapObjects)
	gs := ng - s.s.Goroutines

	// Calculate max players
	maxPlayers := s.o.MaxPlayers
	if s.o.MaxPlayers < pl {
		maxPlayers = pl
	}

	log.Printf("%12d A[%+9d %+9d] %12d HO[%+6d %+6d] %6d GO[%+6d %+6d] %4d PL[%4d]\n",
		m.Alloc, ad, as, m.HeapObjects, hd, hs, ng, gd, gs, pl, maxPlayers,
	)

	s.o.save(m.Alloc, m.HeapObjects, ng, maxPlayers)
}
