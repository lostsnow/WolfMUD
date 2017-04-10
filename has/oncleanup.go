// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package has

// OnCleanup provides a clean up message for a Thing.
//
// Its default implementation is the attr.OnCleanup type.
type OnCleanup interface {
	Attribute

	// CleanupText returns the clean up message for a Thing.
	CleanupText() string
}
