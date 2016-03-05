// Copyright 2015 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package attr

import (
	"code.wolfmud.org/WolfMUD.git/has"
)

// Attribute implements a stub for other attributes. Any types providing
// attributes can embed this type instead of implementing their own Parent and
// SetParent methods.
type Attribute struct {
	parent has.Thing
}

// Some interfaces we want to make sure we implement. If we don't we'll throw
// compile time errors.
//
// TODO: Is it odd Attribute does not implement has.Attribute even though it is
// supposed to be the default implementation?
//var _ has.Attribute = &Attribute{}

// Parent returns the Thing that the Attribute has been added to.
func (a *Attribute) Parent() has.Thing {
	return a.parent
}

// SetParent is used to set the Thing that the Attribute has been added to. If
// it is not currently added to a Thing nil is returned. This method is
// automatically called by the Thing Add method.
func (a *Attribute) SetParent(t has.Thing) {
	a.parent = t
}
