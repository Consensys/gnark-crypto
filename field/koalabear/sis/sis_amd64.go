//go:build !purego

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package sis

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
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
func sisShuffle_avx512(a []koalabear.Element)

//go:noescape
func sisUnshuffle_avx512(a []koalabear.Element)

//go:noescape
func sis512_16_avx512(k256, cosets []koalabear.Element, twiddles [][]koalabear.Element, rag, res []koalabear.Element)
