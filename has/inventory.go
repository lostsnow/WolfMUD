// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

import (
	"sync"
)

type Inventory interface {
	Add(Thing)
	Attribute
	Contents() []Thing
	Count() (int, int, int)
	Crowded() bool
	Found() bool
	List() string
	LockID() uint64
	Remove(Thing) Thing
	Search(string) Thing

	sync.Locker
}
