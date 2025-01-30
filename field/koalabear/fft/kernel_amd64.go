//go:build !purego

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package fft

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"golang.org/x/sys/cpu"
)

var (
	supportAVX512 = cpu.X86.HasAVX512 && cpu.X86.HasAVX512DQ && cpu.X86.HasAVX512VBMI2
)

// q + r'.r = 1, i.e., qInvNeg = - q⁻¹ mod r
// used for Montgomery reduction
const qInvNeg = 2130706431
const q = 2130706433

// index table used in avx512 shuffling
var vInterleaveIndices = []uint64{
	2, 3, 8, 9, 6, 7, 12, 13,
}

//go:noescape
func innerDIFWithTwiddles_avx512(a []koalabear.Element, twiddles []koalabear.Element, start, end, m int)

//go:noescape
func innerDITWithTwiddles_avx512(a []koalabear.Element, twiddles []koalabear.Element, start, end, m int)

func innerDIFWithTwiddles(a []koalabear.Element, twiddles []koalabear.Element, start, end, m int) {
	if !supportAVX512 {
		innerDIFWithTwiddlesGeneric(a, twiddles, start, end, m)
		return
	}
	innerDIFWithTwiddles_avx512(a, twiddles, start, end, m)
}

func innerDITWithTwiddles(a []koalabear.Element, twiddles []koalabear.Element, start, end, m int) {
	if !supportAVX512 {
		innerDITWithTwiddlesGeneric(a, twiddles, start, end, m)
		return
	}
	innerDITWithTwiddles_avx512(a, twiddles, start, end, m)
}

//go:noescape
func kerDIFNP_256_avx512(a []koalabear.Element, twiddles [][]koalabear.Element, stage int)

func kerDIFNP_256(a []koalabear.Element, twiddles [][]koalabear.Element, stage int) {
	if !supportAVX512 {
		kerDIFNP_256generic(a, twiddles, stage)
		return
	}
	kerDIFNP_256_avx512(a, twiddles, stage)
}

//go:noescape
func kerDITNP_256_avx512(a []koalabear.Element, twiddles [][]koalabear.Element, stage int)

func kerDITNP_256(a []koalabear.Element, twiddles [][]koalabear.Element, stage int) {
	if !supportAVX512 {
		kerDITNP_256generic(a, twiddles, stage)
		return
	}
	kerDITNP_256_avx512(a, twiddles, stage)
}
