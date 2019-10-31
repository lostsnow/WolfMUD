// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms_test

import (
	"fmt"
	"testing"

	"code.wolfmud.org/WolfMUD.git/comms"
)

type R struct {
	comms.Ring
	*testing.T
}

func newR(t *testing.T) *R {
	return &R{comms.Ring{}, t}
}

// Function names for use with 'is' method
const (
	Empty = "Empty"
	Full  = "Full"
	Len   = "Len"
	Cap   = "Cap"
	First = "First"
	Last  = "Last"
	Pop   = "Pop"
	Shift = "Shift"
)

// fbool is a lookup table for functions returning a bool
var fbool = map[string]func(*R) bool{
	Empty: (*R).Empty,
	Full:  (*R).Full,
}

// fint is a lookup table for functions returning an int
var fint = map[string]func(*R) int{
	Len: (*R).Len,
	Cap: (*R).Cap,
}

// fint64 is a lookup table for functions returning am int64
var fint64 = map[string]func(*R) int64{
	First: (*R).First,
	Last:  (*R).Last,
	Pop:   (*R).Pop,
	Shift: (*R).Shift,
}

// verify is a helper that compares the content of the ring buffer with an
// equivelent slice.
func (r *R) verify(s []int64) {
	r.Helper()
	if have, want := r.Len(), len(s); have != want {
		r.Errorf("wrong number of elements, have: %d, want: %d", have, want)
		return
	}
	for x := 0; x < r.Len(); x++ {
		if have, want := r.Peek(x), s[x]; have != want {
			r.Errorf("wrong value for element %d, have: %d, want: %d", x, have, want)
		}
	}
	if have, want := r.String(), fmt.Sprintf("%v", s); have != want {
		r.Errorf("invalid dump\nhave: %s\nwant: %s", have, want)
	}
}

// is is a helper function that calls the named function and checks that the
// returned type and value is what is expected.
func (r *R) is(f string, want interface{}) {
	r.Helper()
	// f() bool ?
	if fn, ok := fbool[f]; ok {
		have := fn(r)
		if have != want.(bool) {
			r.Errorf("%s(), have: %t, want: %t", f, have, want)
		}
		return
	}
	// f() int ?
	if fn, ok := fint[f]; ok {
		have := fn(r)
		if have != want.(int) {
			r.Errorf("%s(), have: %d, want: %d", f, have, want)
		}
		return
	}
	// f() int64 ?
	if fn, ok := fint64[f]; ok {
		have := fn(r)
		if have != want.(int64) {
			r.Errorf("%s(), have: %d, want: %d", f, have, want)
		}
		return
	}
	r.Errorf("unknown function: %s()", f)
}

// Test basic conditions on empty buffer
func TestRing_empty(t *testing.T) {
	r := newR(t)
	r.verify([]int64{})
	r.is(Empty, true)
	r.is(Full, false)
	r.is(Len, 0)
	r.is(Cap, 4)
	r.is(First, int64(0))
	r.is(Last, int64(0))
	r.is(Pop, int64(0))
	r.is(Shift, int64(0))
	r.is(Shift, int64(0))
	r.FirstReplace(1)
	r.LastReplace(1)
	r.Popd()
	r.Shiftd()
	r.verify([]int64{})
}

// Test basic conditions on full buffer
func TestRing_full(t *testing.T) {
	r := newR(t)
	for x := 1; !r.Full(); x++ {
		r.Push(int64(x))
	}
	r.verify([]int64{1, 2, 3, 4})
	r.is(Empty, false)
	r.is(Full, true)
	r.is(Len, 4)
	r.is(Cap, 4)
	r.is(First, int64(1))
	r.is(Last, int64(4))
	r.FirstReplace(8)
	r.LastReplace(9)
	r.verify([]int64{8, 2, 3, 9})
	r.is(Shift, int64(8))
	r.is(Pop, int64(9))
	r.verify([]int64{2, 3})
	r.Shiftd()
	r.Popd()
	r.verify([]int64{})
}

// Test basic Push/Pop
func TestRing_pushPop(t *testing.T) {
	r1 := newR(t)
	r2 := newR(t)
	for x := 1; !r1.Full(); x++ {
		r1.Push(int64(x))
	}
	r1.verify([]int64{1, 2, 3, 4})
	r2.verify([]int64{})
	for !r1.Empty() {
		r2.Push(r1.Pop())
	}
	r1.verify([]int64{})
	r2.verify([]int64{4, 3, 2, 1})
}

// Test basic Unshift/Shift
func TestRing_unshiftShift(t *testing.T) {
	r1 := newR(t)
	r2 := newR(t)
	for x := 1; !r1.Full(); x++ {
		r1.Unshift(int64(x))
	}
	r1.verify([]int64{4, 3, 2, 1})
	r2.verify([]int64{})
	for !r1.Empty() {
		r2.Unshift(r1.Shift())
	}
	r1.verify([]int64{})
	r2.verify([]int64{1, 2, 3, 4})
}

// Test basic Poke/Peek
func TestRing_pokePeek(t *testing.T) {
	r := newR(t)
	for x := 1; !r.Full(); x++ {
		r.Push(int64(x))
	}
	for x := 0; x < r.Len(); x++ {
		r.Poke(x, r.Peek(x)+1)
	}
	r.verify([]int64{2, 3, 4, 5})
}

// Test String
func TestRing_string(t *testing.T) {
	r := newR(t)
	for x := 1; !r.Full(); x++ {
		r.Push(int64(x))
	}
	if have, want := r.String(), "[1 2 3 4]"; have != want {
		r.Errorf("String: incorrect output, have: %s want: %q", have, want)
	}

	have := r.String()
	want := fmt.Sprintf("%v", []int64{1, 2, 3, 4})
	if have != want {
		r.Errorf("String: incorrect output, have: %q want: %q", have, want)
	}
}

// Check ring buffer start/end indexes wrap correctly on onverflow of uint8
func TestRing_overflow(t *testing.T) {
	r := newR(t)
	for x := int64(0); x < 1024; x++ {
		if r.Full() {
			r.Shiftd()
		}
		r.Push(x)
		if x > 4 {
			r.verify([]int64{x - 3, x - 2, x - 1, x})
		}
	}
}

// Check ring buffer start/end indexes wrap correctly on underflow of uint8
func TestRing_underflow(t *testing.T) {
	r := newR(t)
	for x := int64(0); x < 1024; x++ {
		if r.Full() {
			r.Popd()
		}
		r.Unshift(x)
		if x > 4 {
			r.verify([]int64{x, x - 1, x - 2, x - 3})
		}
	}
}
