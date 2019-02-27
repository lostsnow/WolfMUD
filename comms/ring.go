// Copyright 2019 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms

import (
	"strconv"
)

// ringSize is the maximum number of elements the Ring buffer can contain. As
// used with connection limiting, in limit.go, this is fixed at 4. This can be
// changed at compile time to 2, 4, 8, 16, 32, 64 or 128 for uint8 indexes.
//
// In general ringSize must be a power of 2 and a maximum equal to the number
// of bits minus 1 for the type size used for the indexes. The maximums are:
//
//   uint8 = 1 <<  7 = 128
//  uint16 = 1 << 15 = 32768
//  uint32 = 1 << 31 = 2147483648
//  uint64 = 1 << 63 = 9223372036854775808
//
// ringSize is currently a uint8 set to 1 << 2 which is 4.
//
// This is not only so that the masking in the 'at' function works but also so
// that when the unsigned index type overflows/underflows there is no
// discontinuity in the indexes.
const ringSize = 1 << 2

// Ring is a very specific implementation of a ring buffer with a fixed
// capacity for use in connection limiting. The implementation is a power of 2
// ring buffer with free running indexes and bit-masking instead of modulo
// arithmetic.
//
// The implementation is complete and can easily be reused and tweaked
// depending on requirements. Copious notes have been provided, but the
// implementation mostly depends on the ringSize, the Ring.elems data type and
// the Ring.start/Ring.end data types.
//
// As is, the ring buffer functions do not return errors for an empty or full
// ring buffer. The ring buffer functions Empty and Full can be used to test
// the ring buffer before calling other functions that would fail.
//
// For performance the ring buffer uses a fixed sized array instead of a slice
// as this eliminates most bounds checking. Also, for performance, the value of
// old elements in the buffer are not zeroed. However, old elements are not
// accessible unless the unexported elems member is accessed directly.
type Ring struct {
	elems [ringSize]int64
	start uint8
	end   uint8
}

// at returns the index into the ring buffer for element at position pos.
func (r *Ring) at(pos uint8) uint8 { return pos & (ringSize - 1) }

// Empty returns true if the ring buffer is empty else false.
func (r *Ring) Empty() bool { return r.start == r.end }

// Full returns true if the ring buffer is full else false.
func (r *Ring) Full() bool { return r.Len() == ringSize }

// Len returns the number of elements currently in use in the ring buffer.
func (r *Ring) Len() int { return int(r.end - r.start) }

// Cap returns the total capacity of the ring buffer.
func (r *Ring) Cap() int { return ringSize }

// Peek returns the value stored in the ring buffer element at position pos. If
// the requested position is invalid, outside of the range 0 <= pos < Len(),
// then the value returned will be 0. If 0 is a valid value to store in the
// ring buffer then the position should be tested before calling Peek so that a
// 0 on error cannot be returned.
func (r *Ring) Peek(pos int) (value int64) {
	if 0 <= pos && pos < r.Len() {
		value = r.elems[r.at(r.start+uint8(pos)+1)]
	}
	return
}

// Poke stores the passed value in the ring buffer element at position pos. The
// update will be ignored if the position is invalid, outside of the range 0 <=
// pos < Len(),
func (r *Ring) Poke(pos int, value int64) {
	if 0 <= pos && pos < r.Len() {
		r.elems[r.at(r.start+uint8(pos)+1)] = value
	}
}

// First returns the value stored in the first element of the ring buffer. If
// the ring buffer is empty 0 will be returned. If 0 is a vlid value to store
// in the ring buffer then First should not be called if Empty returns true.
func (r *Ring) First() (value int64) {
	if !r.Empty() {
		value = r.elems[r.at(r.start+1)]
	}
	return
}

// FirstReplace replaces the value of the first element of the ring buffer with
// the new, provided value. The update will be ignored if the ring buffer is
// empty.
func (r *Ring) FirstReplace(value int64) {
	if !r.Empty() {
		r.elems[r.at(r.start+1)] = value
	}
}

// Last returns the value stored in the last element of the ring buffer. If the
// ring buffer is empty 0 will be returned. If 0 is a vlid value to store in
// the ring buffer then Last should not be called if Empty returns true.
func (r *Ring) Last() (value int64) {
	if !r.Empty() {
		value = r.elems[r.at(r.end)]
	}
	return
}

// LastReplace replaces the value of the last element of the ring buffer with
// the new, provided value. The updated will be ignored if the ring buffer is
// empty.
func (r *Ring) LastReplace(value int64) {
	if !r.Empty() {
		r.elems[r.at(r.end)] = value
	}
}

// Push appends a new element to the end of the ring buffer and sets its value
// to the specified value. If the ring buffer is already full the update will
// be ignored. To avoid silent failures Full can be called before Push to
// make sure the ring buffer is not already full.
func (r *Ring) Push(value int64) {
	if !r.Full() {
		r.end++
		r.elems[r.at(r.end)] = value
	}
}

// Pop removes the last element from the ring buffer and returns its value. If
// the ring buffer is empty then a value of 0 will be returned. If 0 is a vlid
// value to store in the ring buffer then Pop should not be called if Empty
// returns true.
func (r *Ring) Pop() (value int64) {
	if !r.Empty() {
		value = r.elems[r.at(r.end)]
		r.end--
	}
	return
}

// Popd removes and discards the last element from the ring buffer. This is
// more efficient than calling Pop and ignoring the return value.
func (r *Ring) Popd() {
	if !r.Empty() {
		r.end--
	}
}

// Unshift appends a new element to the start of the ring buffer and sets its
// value to the specified value. If the ring buffer is already full the update
// will be ignored. To avoid silent failures Full can be called before Unshift
// to make sure the ring buffer is not already full.
func (r *Ring) Unshift(value int64) {
	if !r.Full() {
		r.elems[r.at(r.start)] = value
		r.start--
	}
}

// Shift removes the first element from the ring buffer and returns its value.
// If the ring buffer is empty then a value of 0 will be returned. If 0 is a
// vlid value to store in the ring buffer then Shift should not be called if
// Empty returns true.
func (r *Ring) Shift() (value int64) {
	if !r.Empty() {
		r.start++
		value = r.elems[r.at(r.start)]
	}
	return
}

// Shiftd removes and discards the first element from the ring buffer. This is
// more efficient than calling Shift and ignoring the return value.
func (r *Ring) Shiftd() {
	if !r.Empty() {
		r.start++
	}
}

// String returns the values of the elements of the ring buffer as a string.
// The values in the string will be in the order of the elements in the ring
// buffer. The format of the returned string is the same as for an int64 slice
// when used with fmt and the $v verb: [value value ...], for easier debugging.
func (r Ring) String() string {
	if r.Len() == 0 {
		return "[]"
	}
	s := []byte{}
	for x := 0; x < r.Len(); x++ {
		v := r.elems[r.at(r.start+uint8(x)+1)]
		s = append(s, ' ')
		s = append(s, strconv.FormatInt(v, 10)...)
	}
	s[0] = '['
	s = append(s, ']')
	return string(s)
}
