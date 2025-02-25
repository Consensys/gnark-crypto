// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fft

import (
	"github.com/consensys/gnark-crypto/field/goldilocks"
)

func innerDIFWithTwiddles(a []goldilocks.Element, twiddles []goldilocks.Element, start, end, m int) {
	innerDIFWithTwiddlesGeneric(a, twiddles, start, end, m)
}

func innerDITWithTwiddles(a []goldilocks.Element, twiddles []goldilocks.Element, start, end, m int) {
	innerDITWithTwiddlesGeneric(a, twiddles, start, end, m)
}

func kerDIFNP_32(a []goldilocks.Element, twiddles [][]goldilocks.Element, stage int) {
	kerDIFNP_32generic(a, twiddles, stage)
}
func kerDITNP_32(a []goldilocks.Element, twiddles [][]goldilocks.Element, stage int) {
	kerDITNP_32generic(a, twiddles, stage)
}

func kerDIFNP_256(a []goldilocks.Element, twiddles [][]goldilocks.Element, stage int) {
	kerDIFNP_256generic(a, twiddles, stage)
}
func kerDITNP_256(a []goldilocks.Element, twiddles [][]goldilocks.Element, stage int) {
	kerDITNP_256generic(a, twiddles, stage)
}
