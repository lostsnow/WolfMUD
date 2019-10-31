// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package log

import (
	"fmt"
	"log"
)

// Conn represents a per connection logging function.
type Conn func(fmt string, v ...interface{})

// NewConn returns a per connection logging function. The function prefixes the
// messages with "[n]" where n is the sequence number of the connection. The
// function is called in exactly the same way as log.Printf in the standard
// library. The output will be written to the standard logger.
func NewConn(seq uint64) Conn {
	s := fmt.Sprintf("[%d] ", seq)
	return func(f string, args ...interface{}) {
		log.Output(2, s+fmt.Sprintf(f, args...))
	}
}
