// Copyright 2014 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar

import (
	"log"
	"strings"
)

// unmarshalers is a map of Unmarshalers keyed by a string 'type name'. Unmarshalers
// for different types call Register to get added to the map. See Register for
// more details.
var unmarshalers map[string]Unmarshaler

// init makes the unmarshalers map so we don't try referencing a nil map which would
// cause a panic.
func init() {
	unmarshalers = make(map[string]Unmarshaler)
}

// RegisterUnmarshaler is used to register an unmarshaler for a type. When a
// Record is unmarshaled it's type attribute is extracted. This is then used as
// the key for looking up the registered umarshaler which is then passed the
// Record for unmarshaling. The name used for the key is uppercased - in effect
// making it case insensitive.
func RegisterUnmarshaler(name string, u Unmarshaler) {
	name = strings.ToUpper(name)
	if _, ok := unmarshalers[name]; !ok {
		unmarshalers[name] = u
	} else {
		panic("Tried to register duplicate unmarshaler: " + name)
	}
	log.Printf("Unmarshaler registered: %T (%s)", u, name)
}
