// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package event implements WolfMUD's asynchronous scripting mechanism.
package event

import (
	"log"
	"math/rand"
	"runtime/debug"
	"time"

	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"
)

// Script is an indirect reference to the cmd.Script function. The cmd package
// cannot be imported directly as it causes a cyclic dependency. However the
// cmd package can import the event package to initialise this variable which
// we can then use. See init function in cmd/state.go for initialization.
var Script func(t has.Thing, input string) string

// FindName is an indirect reference to the attr.FindName function. The attr
// package cannot be imported directly as it causes a cyclic dependency.
// However the attr package can import the event package to initialise this
// variable which we can then use. See init function in attr/name.go for
// initialization.
var FindName func(has.Thing) has.Name

// Cancel is a send only channel that can be used to cancel a queued event.
// When an event is queued via Queue a Cancel channel will be returned. The
// Cancel channel should be closed to cancel the pending event that was queued.
type Cancel chan<- struct{}

// cache is a small pool of reusable timers. As most events fire and then are
// requeued the pool can be quite small and only hold the timers long enough to
// handle the event.
var cache = make(chan *time.Timer, 10)

// Queue schedules a scripted event to happen after the given delay. Queue
// returns a channel to cancel the event and the time of when the event is
// expected to fire.
//
// The firing time is only the expected time and not guaranteed to be 100%
// accurate due to system load, scheduling and processing delays. Typically
// delays are expected to be < 500µs (0.0005s). Delays can be monitored by
// setting Debug.Events to true in the configuration file, in which case the
// delivered messages will contain the amount the event was delayed by, for
// example: "Event delivered (delayed 127.596µs)"
//
// The passed Thing is expected to be the 'actor' for the scripted event.
//
// The input string can use any normal player command and in addition can use
// the scripting only commands starting with the '$' symbol.
//
// The event can be cancelled by closing the returned Cancel channel.
//
// The delay is the period after which the command will be run. The jitter is a
// random amount that can be added to the delay. So the actual delay for an
// event will be between delay and delay+jitter. For a totally random event
// delay can be 0s.
func Queue(thing has.Thing, input string, delay, jitter time.Duration) (Cancel, time.Time) {

	name := FindName(thing).(has.Name).Name("Unknown")

	// Log event notifications?
	logEvents := config.Debug.Events

	cancel := make(chan struct{})

	// Calculate delay in seconds.
	ds := int64(delay / time.Second)

	// Calculate jitter in seconds and pick random jitter
	rj := int64(0)
	if jitter != 0 {
		js := int64(jitter / time.Second)
		if js > 0 {
			rj = rand.Int63n(js)
		}
	}

	// Calc total delay as delay + random jitter in seconds, minimum 1 second
	td := time.Duration(ds+rj) * time.Second
	if td < time.Second {
		td = time.Second
	}

	due := time.Now().Add(td) // Time we expect event to fire

	go func() {

		// If an event goroutine panics try not to bring down the whole server down
		// unless the configuration option Debug.Panic is set to true.
		defer func() {
			if !config.Debug.Panic {
				if err := recover(); err != nil {
					log.Printf("EVENT PANICKED %s: %q Input: %q", thing, name, input)
					log.Printf("%s: %s", err, debug.Stack())
				}
			}
		}()

		var t *time.Timer
		select {
		case t = <-cache:
			t.Reset(td)
		default:
			t = time.NewTimer(td)
		}

		if logEvents {
			log.Printf("Event queued in %s for %s: %q Input: %q", td, thing, name, input)
		}
		select {
		case <-cancel:
			if !t.Stop() {
				<-t.C
				select {
				case cache <- t:
				default:
				}
				return
			}
			if logEvents {
				log.Printf("Event cancelled for %s: %q Input: %q", thing, name, input)
			}
		case <-t.C:
			select {
			case cache <- t:
			default:
			}
			if logEvents {
				log.Printf("Event delivered (delayed %s) for %s: %q Input: %q", time.Now().Sub(due), thing, name, input)
			}
			Script(thing, input)
		}
	}()

	return cancel, due
}
