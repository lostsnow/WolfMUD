// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package test implements some routines useful for testing.
package test

import (
	"testing"
)

// Equal compares two values to see if they are the same. If they are not the
// same an error is reported.
func Equal(t *testing.T, text string, expect, got interface{}) {
	if expect != got {
		t.Errorf("%s expected: %v got: %v", text, expect, got)
	}
}

// NotEqual compares two values to see if they are the same. If they are the
// same an error is reported.
func NotEqual(t *testing.T, text string, expect, got interface{}) {
	if expect == got {
		t.Errorf("%s didn't expect: %v", text, got)
	}
}
