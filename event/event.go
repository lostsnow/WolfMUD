// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// The event package implements WolfMUD's asynchronous scripting mechanism.
package event

import (
	"code.wolfmud.org/WolfMUD.git/config"
	"code.wolfmud.org/WolfMUD.git/has"

	"log"
	"math/rand"
	"runtime/debug"
	"time"
)

// Script is an indirect reference to the cmd.Script function. The cmd package
// cannot be imported directly as it causes a cyclic dependency. However the
// cmd package can import the event package to initialise this variable which
// we can then use. See cmd.Init in cmd/state.go for initialization.
var Script func(t has.Thing, input string) string

// Cancel is a send only channel that can be used to cancel a queued event.
// When an event is queued via Queue a Cancel channel will be returned. The
// Cancel channel should be closed to cancel the pending event that was queued.
type Cancel chan<- struct{}

// Queue schedules a scripted event to happen after the given delay period. The
// command specified in the input will run with access to scripting only
// commands starting with a '$' symbol. The event can be cancelled by closing
// the returned Cancel channel. The thing is the Thing for which the input will
// be scripted. The input is the command to script. The delay is the period
// after which the command will be run. The jitter is a random amount that can
// be added to the delay. So the actual delay for an event will be between
// delay and delay+jitter. For a totally random event delay can be 0s. The
// minimum period+jitter will be 1 second, adjusted if required.
func Queue(thing has.Thing, input string, delay time.Duration, jitter time.Duration) Cancel {

	// Manually find the proper name of the thing instead of using attr.FindName
	// to avoid cyclic imports with the attr and cmd packages.
	name := "Unknown"
	for _, a := range thing.Attrs() {
		if a, ok := a.(has.Name); ok {
			name = a.Name(name)
		}
	}

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
	if td < 1*time.Second {
		td = 1 * time.Second
	}

	// Log event notifications?
	logEvents := config.Debug.Events

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

		t := time.NewTimer(td)
		if logEvents {
			log.Printf("Event queued in %s for %s: %q Input: %q", td, thing, name, input)
		}
		select {
		case <-cancel:
			if logEvents {
				log.Printf("Event cancelled for %s: %q Input: %q", thing, name, input)
			}
			if !t.Stop() {
				<-t.C
			}
		case <-t.C:
			if logEvents {
				log.Printf("Event delivered for %s: %q Input: %q", thing, name, input)
			}
			Script(thing, input)
		}
	}()

	return cancel
}
