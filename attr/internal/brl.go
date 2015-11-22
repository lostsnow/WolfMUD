// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package internal

import (
	"sync"
)

// BRL or 'Big Room Lock' is responsible for all of the 'in game' locking.
// Named in tribute to the Linux kernel 'big kernel lock' that is no more. In
// Linux the BKL stopped concurrency in kernel space. In WolfMUD the BRL stops
// concurrency within a 'location'.  Unlike the BKL the BRL is not a recursive
// lock. If you need to change what you are locking, release all the locks held
// and reacquire them. Otherwise subtle and not so subtle BadThings™ are bound
// to happen.
//
// Each BRL has a unique lock ID associated with it so that locks can be
// obtained and released in a consistent order. This is the classic resource
// hierarchy solution proposed by Dijkstra to the dining philosophers problem
// to avoid deadlocks and livelocks:
//
//  https://en.wikipedia.org/wiki/Dining_philosophers_problem
//
// TODO: Add more details on the BRL, lockID and implications of room level
// locking.
type BRL struct {
	lockID uint64
	sync.Mutex
}

// next is a read only channel used to retrieve the next unique ID number.
var nextLockID <-chan uint64

// init starts a goroutine to generate unique IDs on demand. The goroutine is a
// simple and efficient incrementing counter which blocks and only generates
// the next unique ID when the current one is read using <-nextLockID
func init() {

	// Create bi-directional channel so goroutine can write to it
	// Convert to package level read-only channel
	c := make(chan uint64)
	nextLockID = c

	go func() {
		lockID := uint64(0)
		for {
			c <- lockID
			lockID++
		}
	}()
}

// NewBRL returns an initialised BRL with a unique lock ID.
func NewBRL() BRL {
	return BRL{lockID: <-nextLockID}
}

// LockID returns the unique lock ID associated with a specific BRL.
func (brl BRL) LockID() uint64 {
	return brl.lockID
}
