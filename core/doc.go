// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package core handles all of the main processing of items and player commands
// in WolfMUD. All items are represented via the Thing type with capabilities
// and settings implemented as Thing.Is and Thing.As values. Inventory items
// are held in Thing.In which is a []*Thing. The short names Is, As and In were
// chosen to be meaningful and easily readable in code.
package core
