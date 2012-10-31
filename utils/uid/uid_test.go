// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package uid

import (
	"runtime"
	"testing"
	. "wolfmud.org/utils/test"
)

const (
	LOOPS          = 10
	COUNT_PER_LOOP = 10
	MAX            = LOOPS * COUNT_PER_LOOP
)

func TestSequence(t *testing.T) {
	for x := 0; x < LOOPS; x++ {
		Equal(t, "UID", 1+<-Next, <-Next)
	}
}

func TestConcurrency(t *testing.T) {

	uids := make([]UID, 0)
	results := make(chan UID, MAX)

	// Fire off a number of Goroutines to grab a bunch of UIDs each
	for x := 0; x < LOOPS; x++ {
		go func(results chan UID) {
			for x := 0; x < COUNT_PER_LOOP; x++ {
				results <- <-Next
				runtime.Gosched()
			}
		}(results)
		runtime.Gosched()
	}

	// Wait for results
	for x := 0; x < MAX; x++ {
		temp := <-results
		uids = append(uids, temp)
	}

	// Make sure all results are unique
	for x := 0; x < (MAX - 1); x++ {
		for y := x + 1; y < MAX; y++ {
			NotEqual(t, "Duplicate UID generated", uids[x], uids[y])
		}
	}
}
