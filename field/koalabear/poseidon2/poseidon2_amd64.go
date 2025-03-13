//go:build !purego

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package poseidon2

import (
	fr "github.com/consensys/gnark-crypto/field/koalabear"
)

// q + r'.r = 1, i.e., qInvNeg = - q⁻¹ mod r
// used for Montgomery reduction
const qInvNeg = 2130706431
const q = 2130706433

//go:noescape
func permutation24_avx512(input []fr.Element, roundKeys [][]fr.Element)

//go:noescape
func permutation16_avx512(input []fr.Element, roundKeys [][]fr.Element)

//go:noescape
func permutation16x24_avx512(input *fr.Element, roundKeys [][]fr.Element)

func (h *Permutation) Permutation16x24(input *[16][24]fr.Element) {
	var transposed [24][16]fr.Element
	for i := 0; i < 16; i++ {
		for j := 0; j < 24; j++ {
			transposed[j][i] = input[i][j]
		}
	}
	permutation16x24_avx512(&transposed[0][0], h.params.RoundKeys)

	// do the transpose inverse
	for i := 0; i < 16; i++ {
		for j := 0; j < 24; j++ {
			input[i][j] = transposed[j][i]
		}
	}
}
