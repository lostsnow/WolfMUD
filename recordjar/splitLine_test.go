// Copyright 2018 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar

import (
	"testing"
)

func TestSplitLine(t *testing.T) {
	for _, test := range []struct {
		input string
		field string // Expected field name
		data  string // Expected data value
	}{
		{"a: b", "a", "b"},     // Normal 'field: data'
		{"a:b", "a", "b"},      // 'field:data' - no space
		{"a:", "a", ""},        // field only
		{"a: ", "a", ""},       // field only - with space
		{":b", "", ":b"},       // no field, ':' + data only
		{": b", "", ": b"},     // no field, ': ' + data only
		{"b", "", "b"},         // data only
		{":", "", ":"},         // colon only
		{"", "", ""},           // empty line
		{" ", "", ""},          // space only line
		{"a:b:c", "a", "b:c"},  // field:data + embedded colon
		{"a: b:c", "a", "b:c"}, // field: data + embedded colon

		// Don't expect to see these lines, such lines should be filtered out and
		// not passed to splitLine.
		{"// Comment", "", "// Comment"}, // a comment line
		{"%%", "", "%%"},                 // a record separator
	} {
		t.Run(test.input, func(t *testing.T) {
			have := splitLine.FindSubmatch([]byte(test.input))
			if lhave, lwant := len(have), 3; lhave != lwant {
				t.Errorf("length - have: %d %q, want %d [%q %q %q]",
					lhave, have, lwant, test.input, test.field, test.data)
				return
			}
			if have, want := string(have[1]), test.field; have != want {
				t.Errorf("field - have: %q, want: %q", have, want)
			}
			if have, want := string(have[2]), test.data; have != want {
				t.Errorf("data - have: %q, want: %q", have, want)
			}
		})
	}
}
