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

type mailbox struct {
	queue   chan string // Queued messages waiting to be sent
	lastMsg string      // Last message sent, for de-duplicating
	suffix  string      // Suffix to append to sent messages
}

// mbox stores all of the currenly active mailboxes, indexed by player UID.
var (
	mboxLock sync.RWMutex
	mbox     = make(map[string]*mailbox)
)

// Add a mailbox for the given UID and return a channel for receiving mailbox
// messages.
func Add(uid string) <-chan string {
	b := &mailbox{
		queue: make(chan string, size),
	}
	mboxLock.Lock()
	mbox[uid] = b
	mboxLock.Unlock()
	return b.queue
}

// Delete removes the mailbox for the given UID. Any messages send to a deleted
// mailbox will be discarded. Outstanding messages not retrieved yet will still
// be delivered.
func Delete(uid string) {
	mboxLock.Lock()
	defer mboxLock.Unlock()
	if mbox[uid] != nil {
		close(mbox[uid].queue)
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

// Send writes the given message to the mailbox for the given UID. Priority
// messages are always sent. Non-priority messages are sent if they are not a
// repeat of the last non-priority message sent - this helps cut down on
// message spamming. If mailbox is full then remove + drop oldest message and
// try adding message again.
func Send(uid string, priority bool, msg string) {
	mboxLock.RLock()
	defer mboxLock.RUnlock()

	if mbox[uid] == nil {
		return
	}

	if !priority {
		if mbox[uid].lastMsg == msg {
			return
		}
		mbox[uid].lastMsg = msg
	}

	msg = msg + mbox[uid].suffix

retry:
	select {
	case mbox[uid].queue <- msg:
	default:
		select {
		case <-mbox[uid].queue:
		default:
		}
		goto retry
	}
}

// Suffix sets the current suffix to be appended to sent messages. Setting a
// new suffix only effects new messages and not messages already queued.
func Suffix(uid string, new string) {
	mboxLock.Lock()
	defer mboxLock.Unlock()
	mbox[uid].suffix = new
}
