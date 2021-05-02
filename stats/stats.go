// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package stats implements periodic collection and display of various -
// possibly interesting - statistics. A typical reading might be:
//
//	U[   1Mb  -816b ] O[          1564        +0] G[    39     +0] P 11/11
//
// This shows:
//
//	U[   1Mb  -816b ] - used memory, change since last collection
//	O[  1564      +0] - heap objects, change since last collection
//	G[    39      +0] - Goroutines, change since last collection
//	P 11/11           - Current number of players / maximum number of players
//
// Used memory is rounded to the nearest convenient units: b - bytes, kb -
// kilobytes, Mb - megabytes, Gb - gigabytes, Tb - terabytes, Pb - petabytes,
// Eb - exabytes. Everything above terabytes is included for completeness - but
// 64 unsigned bits *CAN* go up to 15Eb or 18,446,744,073,709,551,615 bytes ;)
//
// The used memory only respresents allocated memory - not memory requested
// from the operating system which is going to be of a larger amount.
//
// The calculations for memory usage have gone through several iterations. The
// current values correspond to the scavenger values for consumed memory seen
// when the server is run with GODEBUG=gctrace=1.
package stats

import (
	"log"
	"runtime"
	"runtime/debug"
	"time"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"
)

var (
	unitPrefixs = [...]string{
		"b", "kb", "Mb", "Gb", "Tb", "Pb", "Eb",
	}
)

// statistics from the last collection
type stats struct {
	m           *runtime.MemStats
	Inuse       uint64
	HeapObjects uint64
	Goroutines  int
	MaxPlayers  int
	ThingCount  uint64
	Allocs      uint64

	u uint64 // current inuse memory
	g int    // current number of goroutines
	p int    // current number of players
	t uint64 // current number of Thing
	a uint64 // current number allocations

	Δu int64 // change in inuse memory
	Δo int   // change in inuse object
	Δg int   // change in number of goroutines
	Δt int64 // change in number of Thing
	Δa int64 // change in number of allocations
}

// Start begins collection and reporting of statistics. The interval between
// updates is controlled via config.StatsRate which if set to zero disables
// collection and reporting.
func Start() {

	if config.Stats.Rate == 0 {
		log.Printf("Stats collection disabled")
		return
	}

	s := &stats{m: &runtime.MemStats{}}
	s.collect() // 1st time initialisation

	go func() {
		for range time.Tick(config.Stats.Rate) {
			s.collect()
		}
	}()

	log.Printf("Stats collection started, frequency: %s", config.Stats.Rate)
}

// collect runs periodically to collect, process and report statistics.
func (s *stats) collect() {

	if config.Stats.GC {
		runtime.GC()
		debug.FreeOSMemory()
	}

	runtime.ReadMemStats(s.m)

	s.u = s.m.HeapInuse + s.m.StackInuse + s.m.MSpanInuse + s.m.MCacheInuse
	s.g = runtime.NumGoroutine()
	s.p = Len()
	s.a = s.m.Mallocs

	s.t = <-attr.ThingCount
	attr.ThingCount <- s.t

	// Calculate difference in resources since last run
	s.Δu = int64(s.u - s.Inuse)
	s.Δo = int(s.m.HeapObjects - s.HeapObjects)
	s.Δg = s.g - s.Goroutines
	s.Δt = int64(s.t - s.ThingCount)
	s.Δa = int64(s.a - s.Allocs)

	// Calculate max players
	maxPlayers := s.MaxPlayers
	if s.MaxPlayers < s.p {
		maxPlayers = s.p
	}

	// Calculate scaled numeric and prefix parts of Inuse and Inuse change
	un, up := uscale(s.u)
	Δun, Δup := scale(s.Δu)

	log.Printf("U[%4d%-2s %+5d%-2s] A[%+9d] O[%14d %+9d] T[%14d %+9d] G[%6d %+6d] P %d/%d",
		un, up, Δun, Δup, s.Δa, s.m.HeapObjects, s.Δo, s.t, s.Δt, s.g, s.Δg, s.p, maxPlayers,
	)

	// Save current stats
	s.Inuse = s.u
	s.HeapObjects = s.m.HeapObjects
	s.Goroutines = s.g
	s.MaxPlayers = maxPlayers
	s.ThingCount = s.t
	s.Allocs = s.a
}

// uscale converts an unsigned number of bytes to a scaled unit of bytes with a
// value less than or equal to 1024 and a unit prefix. For example 42 bytes =
// 42b, 4,242 bytes = 4kb, 424,242 bytes = 414Mb, 42,424,242 bytes = 40Gb.
func uscale(bytes uint64) (scaled int64, scale string) {
	i := 0
	for bytes > 1023 {
		i++
		bytes = bytes >> 10
	}
	return int64(bytes), unitPrefixs[i]
}

// scale converts a signed number of bytes to a scaled unit of bytes with a
// value less than or equal to 1024 and a unit prefix. For example 42 bytes =
// 42b, 4,242 bytes = 4kb, 424,242 bytes = 414Mb, 42,424,242 bytes = 40Gb.
func scale(bytes int64) (scaled int64, scale string) {
	if bytes < 0 {
		scaled, scale = uscale(uint64(bytes * -1))
		scaled *= -1
	} else {
		scaled, scale = uscale(uint64(bytes))
	}
	return
}
