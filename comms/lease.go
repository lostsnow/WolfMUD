// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package comms

import (
	"code.wolfmud.org/WolfMUD.git/config"
)

// lease is used to limit the maximum number of clients that can connect to the
// server at once. Leases should be taken using leaseAcquire and returned using
// leaseRelease.
//
// NOTE: At the moment leases are acquired and released in the client rather
// than the listener. If we put the control into the listener we could get rid
// of the connecting player quicker which would be more advantageous than
// actually setting up a client instance only to tear it down again.
//
// However by having the control in the client we can use the clients logic for
// error handling and data formatting when sending data back to the connecting
// player. This would allow, for example, to send coloured text back. It also
// gives us a mechanism for the future to add maintenance messages, notices and
// other dynamic data instead of a simple "server too busy". It would also
// allow for dynamic control of the leases such as specific open/closing times,
// only allowing beta testers or not allowing new players for a period.
var leases = make(chan struct{}, config.Server.MaxPlayers)

// noLeaseError represents the fact that a lease is currently unavailable.
type noLeaseError struct{}

// Error implements the error interface.
func (noLeaseError) Error() string {
	return "Server Full"
}

// Temporary indicates that a noLeaseError is always a temporary error.
func (noLeaseError) Temporary() bool {
	return true
}

// leaseAcquire is used to try and get a lease for a client. If a lease is
// available calling Error on the client will return nil otherwise Error will
// return noLeaseError. The lease should be released by calling leaseRelease.
func (c *client) leaseAcquire() {
	select {
	case leases <- struct{}{}:
	default:
		c.SetError(noLeaseError{})
	}
}

// leaseRelease is used to release a lease acquired via leaseAcquire. It is
// safe to call leaseRelease even if a call to leaseAcquire failed as the
// client Error will be checked automatically. It is an error to call
// leaseRelease without first having called leaseAcquire.
func (c *client) leaseRelease() {
	if _, ok := c.Error().(noLeaseError); !ok {
		<-leases
	}
}
