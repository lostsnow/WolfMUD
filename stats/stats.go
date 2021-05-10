// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package stats implements periodic collection and display of various -
// possibly interesting - statistics. A typical line from the log might be:
//
//	U[2Mb +1b] A[+48] O[9785 +4] T[265 +0] G[27 +0] W[1.363µs 2.989µs] P[0 0]
//
// The general format of the line is:
//
//	U[n ±n] A[±n] O[n ±n] T[n ±n] G[n ±n] W[n max] P[n max]
//
// The values show the following data:
//
//	U[n  ±n] - used memory, change since last collection
//	A[   ±n] - number of memory allocations since last collection
//	O[n  ±n] - heap objects, change since last collection
//	T[n  ±n] - number of things in the world, change since last collection
//	G[n  ±n] - number of goroutines, change since last collection
//	W[n max] - Lock wait time since last collection / max wait since start
//	P[n max] - Current number of players / maximum number of players
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
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
	"unicode"

	"code.wolfmud.org/WolfMUD.git/attr"
	"code.wolfmud.org/WolfMUD.git/config"
)

// maxLockWait is the maximum time it takes to acquire the locks for processing
// a command. It is reset each time stats are collected. Externally this value
// should be updated by calling the LockWait function which will handle the
// locking.
var (
	maxLockWait      time.Duration
	maxLockWaitMutex sync.RWMutex
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
	LockWait    time.Duration
	MaxLockWait time.Duration
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

	// Get max lock wait time and reset it
	maxLockWaitMutex.Lock()
	s.LockWait = maxLockWait
	maxLockWait = 0
	maxLockWaitMutex.Unlock()
	if s.LockWait > s.MaxLockWait {
		s.MaxLockWait = s.LockWait
	}

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

	log.Printf("U[%4d%-2s %+5d%-2s] A[%+9d] O[%14d %+9d] T[%14d %+9d] G[%6d %+6d] W[%10s %10s] P[%5d %5d]",
		un, up, Δun, Δup, s.Δa, s.m.HeapObjects, s.Δo, s.t, s.Δt, s.g, s.Δg,
		prettyDuration(s.LockWait), prettyDuration(s.MaxLockWait), s.p, maxPlayers,
	)

	// Save current stats
	s.Inuse = s.u
	s.HeapObjects = s.m.HeapObjects
	s.Goroutines = s.g
	s.MaxPlayers = maxPlayers
	s.ThingCount = s.t
	s.Allocs = s.a
}

// LockWait will update the maxLockWait time if required.
func LockWait(wait time.Duration) {

	maxLockWaitMutex.RLock()
	mlw := maxLockWait
	maxLockWaitMutex.RUnlock()

	// Note the need to check the stats didn't change between the initial check
	// and relocking.
	if wait > mlw {
		maxLockWaitMutex.Lock()
		if wait > maxLockWait {
			maxLockWait = wait
		}
		maxLockWaitMutex.Unlock()
	}

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

// prettyDuration formats the given duration to have a precision of 3, with
// trailing zeros. For example: 2.320µs, 397.800ms.
func prettyDuration(d time.Duration) string {
	var D string
	switch {
	case d >= time.Minute:
		D = fmt.Sprintf("%.3fs", d.Seconds())
	case d > time.Second:
		D = d.Round(time.Millisecond).String()
	case d > time.Millisecond:
		D = d.Round(time.Microsecond).String()
	case d > time.Microsecond:
		D = d.Round(time.Nanosecond).String()
	default:
		D = d.String()
	}
	parts := [][]rune{{}, {}, {}}
	idx := 0
	for _, c := range D {
		switch {
		case c == '.':
			idx++
			continue
		case !unicode.IsDigit(c):
			idx = 2
		}
		parts[idx] = append(parts[idx], c)
	}
	parts[1] = append(parts[1], []rune("000")[:3-len(parts[1])]...)
	parts[2] = append(parts[2], []rune("  ")[:2-len(parts[2])]...)
	parts[0] = append(parts[0], '.')
	parts[0] = append(parts[0], parts[1]...)
	parts[0] = append(parts[0], parts[2]...)

	return string(parts[0])
}
