// Copyright 2021 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package mailbox provides asynchronous message delivery to players. A mailbox
// is registered with the Player's UID using the Add function which returns a
// channel for receiving messages. Messages can be sent using the Send function
// with the UID of the recipient player. When the mailbox is no longer required
// Delete should be called to close the channel and remove the mailbox.
package mailbox

import (
	"sync"
)

// size is the maximum number of messages a mailbox can hold before messages
// start being dropped, oldest first.
const size = 100

// Mailbox is where messages/output for players is written to. The key for the
// mailbox is the player's UID.
var mboxLock sync.RWMutex
var mbox = make(map[string]chan string)

// Add a mailbox for the given UID and return a channel for receiving mailbox
// messages.
func Add(uid string) <-chan string {
	q := make(chan string, size)
	mboxLock.Lock()
	mbox[uid] = q
	mboxLock.Unlock()
	return q
}

// Delete removes the mailbox for the given UID. Any messages send to a deleted
// mailbox will be discarded. Outstanding messages not retrieved yet will still
// be delivered.
func Delete(uid string) {
	mboxLock.Lock()
	defer mboxLock.Unlock()
	if mbox[uid] != nil {
		close(mbox[uid])
		delete(mbox, uid)
	}
}

// Len returns the number of mailboxes currently in use.
func Len() (l int) {
	mboxLock.RLock()
	defer mboxLock.RUnlock()
	return len(mbox)
}

// Exists returns true if a mailbox exists for the UID, otherwise false.
func Exists(uid string) bool {
	mboxLock.RLock()
	defer mboxLock.RUnlock()
	return mbox[uid] != nil
}

// Send writes the given message to the mailbox for the given UID. If mailbox
// is full then remove + drop oldest message and try adding message again.
func Send(uid, msg string) {
	mboxLock.RLock()
	defer mboxLock.RUnlock()

	if mbox[uid] == nil {
		return
	}

retry:
	select {
	case mbox[uid] <- msg:
	default:
		select {
		case <-mbox[uid]:
		default:
		}
		goto retry
	}
}
