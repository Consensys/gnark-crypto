//go:build purego || !amd64

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fft

import (
	"github.com/consensys/gnark-crypto/field/babybear"
)

const supportAVX512 = false

func innerDIFWithTwiddles(a []babybear.Element, twiddles []babybear.Element, start, end, m int) {
	innerDIFWithTwiddlesGeneric(a, twiddles, start, end, m)
}

func innerDITWithTwiddles(a []babybear.Element, twiddles []babybear.Element, start, end, m int) {
	innerDITWithTwiddlesGeneric(a, twiddles, start, end, m)
}

func kerDIFNP_256(a []babybear.Element, twiddles [][]babybear.Element, stage int) {
	kerDIFNP_256generic(a, twiddles, stage)
}
func kerDITNP_256(a []babybear.Element, twiddles [][]babybear.Element, stage int) {
	kerDITNP_256generic(a, twiddles, stage)
}
