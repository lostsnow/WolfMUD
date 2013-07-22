// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package sender provides the sender Interface. Senders should format outgoing
// messages. For an implementation see the Client type.
//
// Typically parsers and senders are paired together to process incomming
// (parser) and outgoing (sender) data over a network connection.
package sender

import (
	"code.wolfmud.org/WolfMUD.git/utils/text"
)

// Standard prompt definitions
const (
	PROMPT_NONE    = "\n"
	PROMPT_DEFAULT = text.COLOR_MAGENTA + "\n>"
)

// Interface should be implemented by anything that wants to send data. This
// is typically anything a user will see such as responses to input, menus and
// messages.
type Interface interface {

	// Send is modelled after fmt.Sprintf and takes parameters in the same way.
	// Send should format the message, add any required prompt to the end and
	// then send the message over the network to the connecting client.
	Send(format string, any ...interface{})

	// Sets the currently used prompt
	Prompt(prompt string) (previousPrompt string)
}
