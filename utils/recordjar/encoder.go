// Copyright 2014 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package recordjar

import (
	"strings"
	"time"
)

type Encoder Record

func (e Encoder) String(property, value string) {
	e[property] = strings.TrimSpace(value)
}

func (e Encoder) Keyword(property, value string) {
	e[property] = strings.ToUpper(strings.TrimSpace(value))
}

func (e Encoder) Time(property string, value time.Time) {
	e[property] = value.UTC().Format(time.RFC1123)
}
