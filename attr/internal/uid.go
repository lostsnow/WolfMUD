// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package internal

import (
	"strconv"
	"strings"
)

// UIDPrefix is the prefix to use for unique identifiers.
const UIDPrefix = "#UID-"

// NextUID returns a unique identifier. To keep the identifier compact it is a
// uint that is encoded in base 36 using the digits 0-9 and letters A-Z. This
// will result in an identifier in the range 0 to 3W5E11264SG0G prefixed with
// UIDPrefix. Examples for a unique identifier are #UID-0 & #UID-3W5E11264SG0G.
var NextUID <-chan string

func init() {
	// Create bi-directional channel so goroutine can write to it
	// Convert to package level read-only channel
	c := make(chan string)
	NextUID = c

	go func() {
		UID := uint64(0)
		for {
			c <- UIDPrefix + strings.ToUpper(strconv.FormatUint(UID, 36))
			UID++
		}
	}()
}
