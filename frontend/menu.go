// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package frontend

func (d *Driver) menuDisplay() {
	d.buf.Write([]byte(`
  Main Menu
  ---------

  1. Enter game
  0. Quit

Select an option:`))
	d.nextFunc = d.menuProcess
}

func (d *Driver) menuProcess() {
	if len(d.input) == 0 {
		return
	}
	switch string(d.input) {
	case "1":
		d.gameSetup()
	case "0":
		d.Close()
		d.err = EndOfDataError{}
	default:
		d.buf.WriteString("Invalid option selected.")
	}
}
