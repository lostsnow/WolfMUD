// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

import (
	"sync"
)

type Inventory interface {
	Attribute
	Add(Thing)
	Remove(Thing) Thing
	Search(string) Thing
	Contents() []Thing
	List() string
	LockID() uint64
	Crowded() bool
	Found() bool
	Count() (int, int, int)

	sync.Locker
}
