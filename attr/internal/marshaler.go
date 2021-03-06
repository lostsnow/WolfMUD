// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package internal

import (
	"strings"

	"code.wolfmud.org/WolfMUD.git/has"
)

// Marshalers is a map of registered .wrj marshalers keyed by uppercased field
// names.
var Marshalers = map[string]has.Marshaler{}

// AddMarshaler registers the passed Marshaler as handling marshaling for a
// named field in a .wrj file. The passed Marshaler can be a typed nil pointer
// such as (*Name)(nil).
func AddMarshaler(marshaler has.Marshaler, attr ...string) {
	for _, attr := range attr {
		Marshalers[strings.ToUpper(attr)] = marshaler
	}
}
