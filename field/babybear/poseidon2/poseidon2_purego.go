//go:build purego || !amd64

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package poseidon2

import (
	fr "github.com/consensys/gnark-crypto/field/babybear"
)

func permutation24_avx512(input []fr.Element, roundKeys [][]fr.Element) {
	panic("permutation24_avx512 is not implemented")
}

func permutation16_avx512(input []fr.Element, roundKeys [][]fr.Element) {
	panic("permutation16_avx512 is not implemented")
}

func (h *Permutation) Permutation16x24(input *[16][24]fr.Element) {
	for j := 0; j < 16; j++ {
		h.Permutation(input[j][:])
	}
}
