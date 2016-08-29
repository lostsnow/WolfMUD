// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

func (f *frontend) menuDisplay() {
	f.buf.Write([]byte(`
  Main Menu
  ---------

  1. Enter game
  0. Quit

Select an option:`))
	f.nextFunc = f.menuProcess
}

func (f *frontend) menuProcess() {
	if len(f.input) == 0 {
		return
	}
	switch string(f.input) {
	case "1":
		f.gameInit()
	case "0":
		f.Close()
	default:
		f.buf.WriteString("Invalid option selected.")
	}
}
