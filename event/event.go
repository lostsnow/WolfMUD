// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// The event package implements WolfMUD's asynchronous scripting mechanism.
package event

import (
	"code.wolfmud.org/WolfMUD.git/has"

	"log"
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

// Queue schedules a scripted event to happen after the given delay period.
// The command specified in the input will run wih access to scripting only
// commands starting with a '$' symbol. The event can be cancelled by closing
// the returned Cancel channel.
func Queue(thing has.Thing, input string, delay time.Duration) Cancel {

	// Manually find the proper name of the thing instead of using attr.FindName
	// to avoid cyclic imports with the attr and cmd packages.
	name := "Unknown"
	for _, a := range thing.Attrs() {
		if a, ok := a.(has.Name); ok {
			name = a.Name(name)
		}
	}

	cancel := make(chan struct{})

	go func() {
		t := time.NewTimer(delay)
		log.Printf("Event queued for %q in %s: %s", name, delay, input)
		select {
		case <-cancel:
			log.Printf("Event cancelled for %q: %s", name, input)
			if !t.Stop() {
				<-t.C
			}
		case <-t.C:
			log.Printf("Even delivered for %q: %s", name, input)
			Script(thing, input)
		}
	}()

	return cancel
}
