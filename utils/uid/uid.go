// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package uid provides a unique number generator. To get the next unique number
// simply read from the Next channel:
//
//  MyId := <- uid.Next
//
package uid

import (
	"sync"
)

// UID is currently implemented as a uint64 giving IDs from 1 to
// 18,446,744,073,709,551,615 or 0x1 to 0xFFFFFFFFFFFFFFFF or 18 Quintillion
// IDs also known as 18 exaids. If this is not enough then the type for UID can
// easily be changed. It also means you are probably trying to model every atom
// of your world in WolfMUD or creating a very large galaxy!
type UID uint64

type Interface interface {
	IsAlso(Interface) bool
	UniqueId() UID
}

type UIDLocker interface {
	Interface
	sync.Locker
}

// Next is a read only channel used to retrieve the next ID number.
var Next <-chan UID

// init starts a goroutine to generate IDs on demand. The goroutine function is
// a simple and efficient incrementing counter which blocks on a channel and
// only generates the next ID when the current one is read.
func init() {
	n := make(chan UID) // Create bi-directional channel
	Next = n            // Cast to exported read-only channel
	go func() {
		uid := UID(0)
		for {
			uid++
			n <- uid
		}
	}()
}

// IsAlso tests two UIDs to see if one of them 'is also' the other - hence the
// function's name.
//
// WolfMUD uses a lot of Interfaces and embedded types. So we may be comparing,
// for example, a Player with a Mobile. However this causes issues:
//
// - Mobile and Player are not the same types
// - They can have different interfaces
// - Pointers to a Mobile embedded in a Player will be different (of course)
//
// So to make things easy we have the unique ID and can use either of:
//
// thisPlayer.IsAlso(thisMobile)
// thisPlayer.UniqueId() == thisMobile.UniqueId()
//
// The first example using IsAlso tends to make the code easier to read.
func (u UID) IsAlso(i Interface) bool {
	return u == i.UniqueId()
}

// UniqueId returns the assigned unique ID.
func (u UID) UniqueId() UID {
	return u
}
