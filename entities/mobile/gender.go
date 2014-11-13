// Copyright 2014 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package mobile

// gender defines the non-exported gender type. Only the predefined constants
// are exported to limit the values of gender that can be used.
//
// An uninitialised gender is equivalent to 'It' or indeterminable. This may not
// sound 'politically correct' but a lot of monsters are just 'It'.
type gender int

// Constants for exported gender types. These are the only gender types that are valid.
const (
	GenderIt gender = iota
	GenderMale
	GenderFemale
)

// Gender title cased lookup table. When using the various helpers and building
// texts it avoids repetative importing of the strings package just to use
// Title.
var imfTC = [3][4]string{{"It", "He", "She"}, {"Its", "His", "Hers"}, {"It", "Male", "Female"}}

// Gender lower cased lookup table. When using the various helpers and building
// texts it avoids repetative importing of the strings package just to use
// Title.
var imfLC = [3][4]string{{"it", "he", "she"}, {"its", "his", "hers"}, {"it", "male", "female"}}

// Gender returns the current gender which will be one of the exported constants.
func (g gender) Gender() gender {
	return g
}

// ItHeShe returns title cased "It", "He" or "She" depending on the current
// gender.
func (g gender) ItHeShe() string {
	return imfTC[0][g]
}

// LCItHeShe returns lower cased "it", "he" or "she" depending on the current
// gender.
func (g gender) LCItHeShe() string {
	return imfLC[0][g]
}

// ItsHisHers returns title cased "Its", "His" or "Hers" depending on the
// current gender.
func (g gender) ItsHisHers() string {
	return imfTC[1][g]
}

// LCItsHisHers returns lower cased "its", "his" or "hers" depending on the
// current gender.
func (g gender) LCItsHisHers() string {
	return imfLC[1][g]
}

// ItMaleFemale returns title cased "It", "Male" or "Female" depending on the
// current gender.
func (g gender) ItMaleFemale() string {
	return imfTC[2][g]
}

// LCItMaleFemale returns lower cased "it", "male" or "female" depending on the
// current gender.
func (g gender) LCItMaleFemale() string {
	return imfLC[2][g]
}
