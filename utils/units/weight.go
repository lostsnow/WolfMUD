// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package units

import (
	"bytes"
	"strconv"
)

// Weight defines the standard units of weight. For a more modern game
// different units of weight can easily be defined.
type Weight ounces

// A few useful weight constants
const (
	ZERO_WEIGHT  = ounces(0)
	POUND        = ounces(16)
	HALF_POUND   = POUND / 2
)

// String implements the Stringer interface so that a weight can have a meaning
// description like "1lb and 2oz" just by using %s and %v etc.
func (w Weight) String() string {
	return ounces(w).String()
}

func (w Weight) Int() int {
	return int(w)
}

// ounces is used as the standard weight for items.
type ounces int

// String displays ounces as pounds and ounces. Ounces are only displayed for
// light weights. If the weight is 2 pounds or more then the ounces are not
// displayed but the pounds are rounded up if there are over 8 ounces. For
// example:
//
//	2 ounces is displayed as "2oz"
//	18 ounces displays as "1lb and 2oz"
//	88 ounces displays as "5lb" and not "5lb and 8oz"
//	89 ounces displays as "6lb" and not "5lb and 9oz"
//
func (o ounces) String() string {
	switch o {
		case ZERO_WEIGHT:
			return "nothing";
		case HALF_POUND:
			return "half a pound"
		case POUND:
			return "a pound"
	}

	b := new(bytes.Buffer)

	oz := o % POUND
	lb := (o - oz) / POUND

	if lb >= 2 && oz > HALF_POUND {
		lb++
	}

	if lb != ZERO_WEIGHT {
		b.WriteString(strconv.Itoa(int(lb)))
		b.WriteString("lb")
		if oz != ZERO_WEIGHT && lb < 2 {
			b.WriteString(" and ")
		}
	}

	if oz != ZERO_WEIGHT && lb < 2 {
		b.WriteString(strconv.Itoa(int(oz)))
		b.WriteString("oz")
	}

	return b.String()
}
