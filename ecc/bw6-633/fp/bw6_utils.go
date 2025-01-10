// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fp

// MulByNonResidue multiplies a fp.Element by 2
func (z *Element) MulByNonResidue(x *Element) *Element {
	z.Double(x)
	return z
}
