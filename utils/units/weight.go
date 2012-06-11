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

func (w Weight) String() string {
	return ounces(w).String()
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
	b := new(bytes.Buffer)

	o_int := int(o)

	oz := o_int % 16
	lb := (o_int - oz) / 16

	if lb >= 2 && oz > 8 {
		lb++
	}

	if lb != 0 {
		b.WriteString(strconv.Itoa(lb))
		b.WriteString("lb")
	}
	if oz != 0 && lb < 2 {
		if b.Len() != 0 {
			b.WriteString(" and ")
		}
		b.WriteString(strconv.Itoa(oz))
		b.WriteString("oz")
	}

	return b.String()
}
