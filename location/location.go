/*
	Copyright 2012 Andrew 'Diddymus' Rolfe. All rights reserved.
*/

/*
	Package location implements all of the different location types available
	in WolfMUD.
*/
package location

/*
	Location is an interface for different location types.
*/
type Location interface {
	Exits()
	Look()
	Move(direction) Location
	SetExit(direction, Location)
}
