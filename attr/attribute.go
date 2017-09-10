// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"log"
	"sync"

	"code.wolfmud.org/WolfMUD.git/has"
)

// Attribute implements a stub for other attributes. Any types providing
// attributes can embed this type instead of implementing their own Parent and
// SetParent methods. Updating and querying the parent is concurrent safe.
//
// NOTE: Attribute does NOT provide a default Copy implementation. Each
// attribute must implement its own Copy method. This is due to the fact that
// other attributes will know best how to create copies based on their own
// implementation.
type Attribute struct {
	rwmutex sync.RWMutex
	parent  has.Thing
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
//
// TODO: Is it odd Attribute does not implement has.Attribute even though it is
// supposed to be the default implementation?
//var _ has.Attribute = &Attribute{}

// Parent returns the Thing that the Attribute has been added to.
func (a *Attribute) Parent() has.Thing {
	a.rwmutex.RLock()
	defer a.rwmutex.RUnlock()
	return a.parent
}

// SetParent is used to set the Thing that the Attribute has been added to. If
// it is not currently added to a Thing nil is returned. This method is
// automatically called by the Thing Add method.
func (a *Attribute) SetParent(t has.Thing) {
	a.rwmutex.Lock()
	a.parent = t
	a.rwmutex.Unlock()
}

// FOR DEVELOPMENT ONLY SO WE DON'T HAVE TO IMPLEMENT Marshal ON ALL THE
// ATTRIBUTES AT ONCE. REMOVE AS SOON AS ALL ATTRIBUTES UPDATED.
func (a *Attribute) Marshal(attr has.Attribute) []byte {
	log.Println("[DEBUG] dummy marshal")
	return []byte{}
}

// Free makes sure references are nil'ed when the Attribute is freed. Other
// attributes should override Free to release their own references and
// resources. Attributes that implement their own Free method should also call
// Attribute.Free.
func (a *Attribute) Free() {
	if a != nil {
		a.rwmutex.Lock()
		a.parent = nil
		a.rwmutex.Unlock()
	}
}
