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
	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"

	"log"
	"runtime"
	"runtime/debug"
	"time"
)

var (
	unitPrefixs = [...]string{
		"b", "kb", "Mb", "Gb", "Tb", "Pb", "Eb",
	}
)

// statistics from the last collection
type stats struct {
	Inuse       uint64
	HeapObjects uint64
	Goroutines  int
	MaxPlayers  int
	ThingCount  uint64
}

// Start begins collection and reporting of statistics. The interval between
// updates is controlled via config.StatsRate which if set to zero disables
// collection and reporting.
func Start() {

	if config.Stats.Rate == 0 {
		log.Printf("Stats collection disabled")
		return
	}

	s := &stats{}
	s.collect() // 1st time initialisation

	go func() {
		for _ = range time.Tick(config.Stats.Rate) {
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

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	u := m.HeapInuse + m.StackInuse + m.MSpanInuse + m.MCacheInuse
	g := runtime.NumGoroutine()
	p := Len()

	t := <-attr.ThingCount
	attr.ThingCount <- t

	// Calculate difference in resources since last run
	Δu := int64(u - s.Inuse)
	Δo := int(m.HeapObjects - s.HeapObjects)
	Δg := g - s.Goroutines
	Δt := int64(t - s.ThingCount)

	// Calculate max players
	maxPlayers := s.MaxPlayers
	if s.MaxPlayers < p {
		maxPlayers = p
	}

	// Calculate scaled numeric and prefix parts of Inuse and Inuse change
	un, up := uscale(u)
	Δun, Δup := scale(Δu)

	log.Printf("U[%4d%-2s %+5d%-2s] O[%14d %+9d] T[%14d %+9d] G[%6d %+6d] P %d/%d",
		un, up, Δun, Δup, m.HeapObjects, Δo, t, Δt, g, Δg, p, maxPlayers,
	)

	// Save current stats
	s.Inuse = u
	s.HeapObjects = m.HeapObjects
	s.Goroutines = g
	s.MaxPlayers = maxPlayers
	s.ThingCount = t
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
