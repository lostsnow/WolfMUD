// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// Marshaler provides the ability to marshal and unmarshal fields for a .wrj
// (WolfMUD Record Jar) file to and from Attributes. It should be implemnted by
// all Attribute types.
type Marshaler interface {

	// Unmarshal takes the []byte and returns an Attribute for the []byte data.
	// If the returned Attribute is an untyped nil it should be ignored.
	Unmarshal([]byte) Attribute

	// Marshal returns a tag and []byte that represents an Attribute.
	Marshal() (tag string, data []byte)
}
