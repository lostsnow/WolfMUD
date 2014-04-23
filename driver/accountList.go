// Copyright 2013 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package driver

import (
	"errors"
	"log"
)

// accounts is a channel that should buffer a single map of logged in accounts
// keyed by accountId. The map then acts as both data and lock.  To access the
// accounts you take the lock by removing the map, use it, then put the map
// back into the channel to release the lock. While the map is in use other go
// routines will block until the map is put back and can be read again. As maps
// are reference types only a reference should actually go into the channel
// keeping things lightweight.
var accounts chan map[string]struct{}

// init sets up the account tracking map and channel
func init() {
	accounts = make(chan map[string]struct{}, 1)
	accounts <- make(map[string]struct{})
}

func (d *driver) login() error {
	a := <-accounts
	defer func() { accounts <- a }()

	if _, found := a[d.account]; found {
		log.Printf("Duplicate login: %s", d.sender)
		return errors.New("Duplicate login")
	}

	log.Printf("Successful login: %s", d.sender)
	a[d.account] = struct{}{}

	return nil
}

func (d *driver) Logout() {
	a := <-accounts
	defer func() { accounts <- a }()

	// Check if we are already logged out and save time...
	if _, found := a[d.account]; !found {
		return
	}

	if d.player != nil && d.player.Locate() != nil {
		d.player.Parse("QUIT")
	}

	log.Printf("Logout: %s", d.sender)
	delete(a, d.account)
}
