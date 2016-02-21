// Copyright 2016 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

// Package has provides interfaces for Attributes. All interfaces defined in
// this package should use either standard Go types or other interfaces. This
// package should not import other packages, especially the attr package. This
// helps in avoiding cyclic imports.
//
// For each interface defined in the has package there is usually a default
// implementation of it in the attr package with the same name. For example
// has.Inventory has the implementation attr.Inventory.
//
// This package does not use the 'er' naming convention for interfaces such as
// Reader and Writer. It uses the actual attribute names such as Name and
// Description. This makes sense for WolfMUD when attributes are qualified with
// the has package name: has.Name, has.Description, has.Inventory.
package has
