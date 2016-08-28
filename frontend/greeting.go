// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

import (
	"code.wolfmud.org/WolfMUD.git/config"
)

// greetingDisplay shows the welcome message stored in the server configuration
// file free text area. For more information see the config package.
func (f *frontend) greetingDisplay() {
	f.buf.Write(config.Server.Greeting)
	f.accountDisplay()
}
