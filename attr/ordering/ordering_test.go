// Copyright 2020 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package ordering_test

import (
	"strings"
	"testing"

	_ "code.wolfmud.org/WolfMUD.git/attr" // Register marshalers
	"code.wolfmud.org/WolfMUD.git/attr/internal"
	"code.wolfmud.org/WolfMUD.git/attr/ordering"
)

// TestOrdering_present makes sure all marshable attributes are accounted for
// in the attribute ordering.
func TestOrdering_present(t *testing.T) {

	// Attribute not current persisted in record jars
	notPersisted := map[string]struct{}{
		"PLAYER": struct{}{},
		"LOCATE": struct{}{},
	}

	o := make(map[string]struct{})
	for _, field := range ordering.Attributes {
		o[strings.ToUpper(field)] = struct{}{}
	}

	for m := range internal.Marshalers {
		if _, ok := notPersisted[m]; ok {
			continue
		}
		if _, ok := o[m]; !ok {
			t.Errorf("Attribute not in Ordering: %q", m)
		}
	}
}
