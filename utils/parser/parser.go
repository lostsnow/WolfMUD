// Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package parser provides the parser Interface. Parsers should interpret
// incoming data. They can implement logins, menu systems, mini chat systems or
// player sessions.
//
// Typically parsers and senders are paired together to process incomming
// (parser) and outgoing (sender) data over a network connection.
package parser

// Interface should be implemented by anything that wants to interpret incoming
// data. For an example implementation see the Player type which interprets
// commands from the Client type. Typically a parser returns responses via a
// sender.
type Interface interface {

	// Parse should act on the passed input which could be a selected menu
	// option, a chat message to send or a players command.
	Parse(input string)

	// Name should return the name associated with the parser, usually a player's
	// name but it can also be a chat name, the name of a menu or an empty
	// string. It is used by the Client type for debugging messages and should
	// also be implemented by Thing types anyway.
	Name() string

	// Destroy should cleanly shut down the parser and release any resources when
	// called.
	Destroy()

	// If the parser is quitting - exit selected in a menu, player issues quit
	// command - this should return true otherwise false.
	IsQuitting() bool

	// Next returns the next parser to be used. This allows us to transfer
	// control from one parser to another. For example from login→player→login ad
	// infinitum. It also allows for modular parsers.
	Next() Interface
}
