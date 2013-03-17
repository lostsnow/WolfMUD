// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package stats implements periodic collection and display of various -
// possibly interisting - statistics. A typical reading might be:
//
//	A[   1Mb  -816b ] HO[          1564        +0] GO[    39     +0] PL 11/11
//
// This shows:
//
//	 A[   1Mb  -816b ] - allocated memory, change since last collection
//	HO[  1564      +0] - heap objects, change since last collection
//	GO[    39      +0] - Goroutines, change since last collection
//	PL 11/11           - Current number of players / maximum number of players
//
// Allocated memory is rounded to the nearest convenient units: b - bytes, kb -
// kilobytes, Mb - megabytes, Gb - gigabytes, Tb - terabytes, Pb - petabytes,
// Eb - exabytes. Everything above terabytes is included for completeness - but
// 64 unsigned bits *CAN* go up to 15Eb or 18,446,744,073,709,551,615 bytes ;)
package stats

import (
	"code.wolfmud.org/WolfMUD.git/entities/mobile/player"
	"log"
	"runtime"
	"time"
)

// Interval can be changed before calling Start to set a different collection
// interval.
//
// TODO: When we have sorted out global settings Interval needs moving there.
var (
	Interval    = 5 * time.Minute // Time  between collections
	unitPrefixs = [...]string{
		"b", "kb", "Mb", "Gb", "Tb", "Pb", "Eb",
	}
)

// statistics from the last collection
type stats struct {
	Alloc       uint64
	HeapObjects uint64
	Goroutines  int
	MaxPlayers  int
}

// Start begins collection and reporting of statistics. The interval between
// updates is controlled via the stats.Interval variable and can be set before
// calling Start.
func Start() {
	s := &stats{}
	s.collect() // 1st time initialisation

	go func() {
		for _ = range time.Tick(Interval) {
			s.collect()
		}
	}()

	log.Printf("Stats collection started, frequency: %s", Interval)
}

// collect runs periodically to collect, process and report statistics.
func (s *stats) collect() {
	runtime.GC()
	runtime.Gosched()

	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	g := runtime.NumGoroutine()
	p := player.PlayerList.Length()

	// Calculate difference in resources since last run
	Δa := int64(m.Alloc - s.Alloc)
	Δh := int(m.HeapObjects - s.HeapObjects)
	Δg := g - s.Goroutines

	// Calculate max players
	maxPlayers := s.MaxPlayers
	if s.MaxPlayers < p {
		maxPlayers = p
	}

	// Calculate scaled numeric and prefix parts of Alloc and Alloc difference
	an, ap := uscale(m.Alloc)
	Δan, Δap := scale(Δa)

	log.Printf("A[%4d%-2s %+5d%-2s] HO[%14d %+9d] GO[%6d %+6d] PL %d/%d\n",
		an, ap, Δan, Δap, m.HeapObjects, Δh, g, Δg, p, maxPlayers,
	)

	// Save current stats
	s.Alloc = m.Alloc
	s.HeapObjects = m.HeapObjects
	s.Goroutines = g
	s.MaxPlayers = maxPlayers
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
