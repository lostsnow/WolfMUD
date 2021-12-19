// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package stats is used to report on runtime statistics for the WolfMUD
// server. The package is configured by the Stats.Rate and Stats.GC values in
// the server configuration file. Statistics will be written to the server log
// in the following format:
//
//	A[n] O[n ±n] T[n ±n] E[n ±n] P[n max]
//
// Where:
//
//	A[    n] - Runtime allocations since last collection
//	O[n  ±n] - Runtime objects / change since last collection
//	T[n  ±n] - Thing in the world / change since last collection
//	E[n  ±n] - In-flight active events / change since last collection
//	P[n max] - Current number of players / maximum number of players
//
// Forcing garbage collection will provide more accurate statistics. However,
// there will be a nominal hit on performance.
package stats

import (
	"log"
	"runtime"
	"time"

	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/core"
	"code.wolfmud.org/WolfMUD.git/mailbox"
)

// stats is used to record the previous and current statistics
type stats struct {
	allocs   uint64 // Total runtime allocations
	objects  uint64 // Total runtime objects
	Mbox     int    // Number of active mailboxes/players
	maxMbox  int    // Maximum mailboxes/players
	Thing    int    // Number of Thing in world
	maxThing int    // Maximum Things in world
	Event    int    // Number of in-flight/active events
}

type pkgConfig struct {
	rate time.Duration
	gc   bool
}

// cfg setup by Config and should be treated as immutable and not changed.
var cfg pkgConfig

// Config sets up package configuration for settings that can't be constants.
// It should be called by main, only once, before anything else starts. Once
// the configuration is set it should be treated as immutable an not changed.
func Config(c config.Config) {
	cfg = pkgConfig{
		rate: c.Stats.Rate,
		gc:   c.Stats.GC,
	}
}

// Start statistics collection. The configured collection rate is checked, if
// zero statistics collection is disabled. Otherwise statistics are collected
// by calling stats.collect every cfg.rate periods.
func Start() {
	if cfg.rate == 0 {
		log.Print("Statistics collection disabled")
		return
	}

	log.Printf("Started logging statistics, frequency: %s, GC: %t",
		cfg.rate, cfg.gc,
	)
	t := time.NewTicker(cfg.rate)
	go collect(t)
}

// collect gathers, records and displays statistics every time the passed
// Ticker channel fires.
func collect(t *time.Ticker) {
	var (
		m  = &runtime.MemStats{} // runtime statistics
		p  = &stats{}            // Previously collected statistics
		c  = &stats{}            // Currently collected statistics
		Δa int64                 // Change in allocation count
		Δo int64                 // Change in object count
		Δt int64                 // Change in Thing count
		Δe int64                 // Change in event count
	)
	p.store(m)
	for {
		select {
		case <-t.C:
			c.store(m)
			Δa = int64(c.allocs - p.allocs)
			Δo = int64(c.objects - p.objects)
			Δt = int64(c.Thing - p.Thing)
			Δe = int64(c.Event - p.Event)
			log.Printf("A[%9d] O[%9d %+7d] T[%9d %+6d] E[%9d %+6d] P[%7d %7d]",
				Δa, c.objects, Δo, c.Thing, Δt, c.Event, Δe, c.Mbox, c.maxMbox)
			p, c = c, p
		}
	}
}

// store records the current statistics in the receiver. The passed
// runtime.MemStats allows a preallocated instance to be reused. If cfg.gc is
// true then garbage collection will be forced before reading the runtime
// memory statistics.
func (s *stats) store(m *runtime.MemStats) {
	if cfg.gc {
		runtime.GC()
	}
	runtime.ReadMemStats(m)
	s.allocs = m.TotalAlloc
	s.objects = m.Mallocs - m.Frees
	s.Mbox = mailbox.Len()
	if s.Mbox > s.maxMbox {
		s.maxMbox = s.Mbox
	}
	s.Thing = core.ThingCount()
	if s.Thing > s.maxThing {
		s.maxThing = s.Thing
	}
	s.Event = core.EventCount()
}
