//go:build !purego

// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package poseidon2

import (
	fr "github.com/consensys/gnark-crypto/field/babybear"
)

// q + r'.r = 1, i.e., qInvNeg = - q⁻¹ mod r
// used for Montgomery reduction
const qInvNeg = 2013265919
const q = 2013265921

// index table used in avx512 shuffling
var vInterleaveIndices = []uint64{
	1, 0, 3, 2, 5, 4, 7, 6,
}

var vInterleaveIndices2 = []uint32{
	4, 5, 6, 7, 0, 1, 2, 3,
}

var vM4Indices = []uint64{
	1, 2, 3, 0, 5, 6, 7, 4,
}

//go:noescape
func permutation24_avx512(input []fr.Element, roundKeys [][]fr.Element)

const width = 24

func Permutation24_avx512(input []fr.Element, roundKeys [][]fr.Element) {
	permutation24_avx512(input, roundKeys)

	// matMulExternalInPlace2(input)

	// const fullRounds = 6
	// const partialRounds = 21
	// const rf = fullRounds / 2
	// for i := 0; i < rf; i++ {
	// 	// one round = matMulExternal(sBox_Full(addRoundKey))
	// 	addRoundKeyInPlace(i, input, roundKeys)
	// 	for j := 0; j < width; j++ {
	// 		sBox(j, input)
	// 	}
	// 	matMulExternalInPlace(input)
	// }

	// for i := rf; i < rf+partialRounds; i++ {
	// 	// one round = matMulInternal(sBox_sparse(addRoundKey))
	// 	addRoundKeyInPlace(i, input, roundKeys)
	// 	sBox(0, input)
	// 	matMulInternalInPlace(input)
	// }
	// for i := rf + partialRounds; i < fullRounds +partialRounds; i++ {
	// 	// one round = matMulExternal(sBox_Full(addRoundKey))
	// 	addRoundKeyInPlace(i, input, roundKeys)
	// 	for j := 0; j < width; j++ {
	// 		sBox(j, input)
	// 	}
	// 	matMulExternalInPlace(input)
	// }

}

func sBox(index int, input []fr.Element) {
	var tmp fr.Element
	tmp.Set(&input[index])
	// sbox degree is 3
	input[index].Square(&input[index]).
		Mul(&input[index], &tmp)
}

func matMulInternalInPlace(input fr.Vector) {
	var sum fr.Element
	sum.Set(&input[0])
	for i := 1; i < width; i++ {
		sum.Add(&sum, &input[i])
	}
	vDiag24 := fr.Vector(diag24[:])
	input.Mul(input, vDiag24)
	for i := range input {
		input[i].Add(&sum, &input[i])
	}
}

func addRoundKeyInPlace(round int, input []fr.Element, roundKeys [][]fr.Element) {
	for i := 0; i < len(roundKeys[round]); i++ {
		input[i].Add(&input[i], &roundKeys[round][i])
	}
}

func matMulM4InPlace(s []fr.Element) {
	c := len(s) / 4
	for i := 0; i < c; i++ {
		var t01, t23, t0123, t01123, t01233 fr.Element
		t01.Add(&s[4*i], &s[4*i+1])
		t23.Add(&s[4*i+2], &s[4*i+3])
		t0123.Add(&t01, &t23)
		t01123.Add(&t0123, &s[4*i+1])
		t01233.Add(&t0123, &s[4*i+3])
		// The order here is important. Need to overwrite x[0] and x[2] after x[1] and x[3].
		s[4*i+3].Double(&s[4*i]).Add(&s[4*i+3], &t01233)
		s[4*i+1].Double(&s[4*i+2]).Add(&s[4*i+1], &t01123)
		s[4*i].Add(&t01, &t01123)
		s[4*i+2].Add(&t23, &t01233)
	}
}

func matMulExternalInPlace(input []fr.Element) {
	// at this stage t is supposed to be a multiple of 4
	// the MDS matrix is circ(2M4,M4,..,M4)
	// permutation24_avx512(input, nil)
	matMulM4InPlace(input)
	tmp := make([]fr.Element, 4)
	for i := 0; i < width/4; i++ {
		tmp[0].Add(&tmp[0], &input[4*i])
		tmp[1].Add(&tmp[1], &input[4*i+1])
		tmp[2].Add(&tmp[2], &input[4*i+2])
		tmp[3].Add(&tmp[3], &input[4*i+3])
	}
	for i := 0; i < width/4; i++ {
		input[4*i].Add(&input[4*i], &tmp[0])
		input[4*i+1].Add(&input[4*i+1], &tmp[1])
		input[4*i+2].Add(&input[4*i+2], &tmp[2])
		input[4*i+3].Add(&input[4*i+3], &tmp[3])
	}
}

func matMulExternalInPlace2(input []fr.Element) {
	// at this stage t is supposed to be a multiple of 4
	// the MDS matrix is circ(2M4,M4,..,M4)
	// permutation24_avx512(input, nil)
	// matMulM4InPlace(input)
	tmp := make([]fr.Element, 4)
	for i := 0; i < width/4; i++ {
		tmp[0].Add(&tmp[0], &input[4*i])
		tmp[1].Add(&tmp[1], &input[4*i+1])
		tmp[2].Add(&tmp[2], &input[4*i+2])
		tmp[3].Add(&tmp[3], &input[4*i+3])
	}
	for i := 0; i < width/4; i++ {
		input[4*i].Add(&input[4*i], &tmp[0])
		input[4*i+1].Add(&input[4*i+1], &tmp[1])
		input[4*i+2].Add(&input[4*i+2], &tmp[2])
		input[4*i+3].Add(&input[4*i+3], &tmp[3])
	}
}
