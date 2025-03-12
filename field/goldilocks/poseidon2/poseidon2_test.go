// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package poseidon2

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/goldilocks"
)

func TestMulMulInternalInPlaceWidth8(t *testing.T) {
	var input, expected [8]fr.Element
	for i := range input {
		input[i].SetRandom()
	}

	expected = input

	h := NewPermutation(8, 6, 17)
	h.matMulInternalInPlace(expected[:])

	var sum fr.Element
	sum.Set(&input[0])
	for i := 1; i < h.params.Width; i++ {
		sum.Add(&sum, &input[i])
	}
	for i := 0; i < h.params.Width; i++ {
		input[i].Mul(&input[i], &diag8[i]).
			Add(&input[i], &sum)
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mat mul internal w/ diagonal doesn't match hand calculated")
		}
	}
}

func TestMulMulInternalInPlaceWidth12(t *testing.T) {
	var input, expected [12]fr.Element
	for i := range input {
		input[i].SetRandom()
	}

	expected = input

	h := NewPermutation(12, 6, 17)
	h.matMulInternalInPlace(expected[:])

	var sum fr.Element
	sum.Set(&input[0])
	for i := 1; i < h.params.Width; i++ {
		sum.Add(&sum, &input[i])
	}
	for i := 0; i < h.params.Width; i++ {
		input[i].Mul(&input[i], &diag12[i]).
			Add(&input[i], &sum)
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mat mul internal w/ diagonal doesn't match hand calculated")
		}
	}
}
func TestPoseidon2Width8(t *testing.T) {
	var input, expected [8]fr.Element
	// these are random values generated by SetRandom()
	input[0].SetUint64(16906858123866173649)
	input[1].SetUint64(15166437626912738600)
	input[2].SetUint64(5043155767520437527)
	input[3].SetUint64(4803372521910203894)
	input[4].SetUint64(1363381407771951133)
	input[5].SetUint64(14358392110422722767)
	input[6].SetUint64(16147940662011238603)
	input[7].SetUint64(17042226261559028170)

	expected[0].SetUint64(9592598718001559987)
	expected[1].SetUint64(3706879638445770744)
	expected[2].SetUint64(17276696801585841081)
	expected[3].SetUint64(4798871633124733906)
	expected[4].SetUint64(13363852300480597050)
	expected[5].SetUint64(17026630749095291654)
	expected[6].SetUint64(16473007323551129424)
	expected[7].SetUint64(10515428028369692011)

	h := NewPermutation(8, 6, 17)
	h.Permutation(input[:])
	for i := 0; i < h.params.Width; i++ {
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mismatch error")
		}
	}
}

func TestPoseidon2Width12(t *testing.T) {
	var input, expected [12]fr.Element
	// these are random values generated by SetRandom()
	input[0].SetUint64(16177261168397522151)
	input[1].SetUint64(17965107813155799464)
	input[2].SetUint64(4862396544291584838)
	input[3].SetUint64(3316843815481829987)
	input[4].SetUint64(5261586417311804404)
	input[5].SetUint64(10778243380389816710)
	input[6].SetUint64(7667572003603320753)
	input[7].SetUint64(2325393195433953062)
	input[8].SetUint64(2060868681750658110)
	input[9].SetUint64(2254293530099160974)
	input[10].SetUint64(6150660266886974089)
	input[11].SetUint64(14161738010109367755)

	expected[0].SetUint64(14366152479958620597)
	expected[1].SetUint64(6220113587113887785)
	expected[2].SetUint64(14300084842296079345)
	expected[3].SetUint64(8434700876601154441)
	expected[4].SetUint64(13811271242031833355)
	expected[5].SetUint64(10611066669541572840)
	expected[6].SetUint64(7885561287590750763)
	expected[7].SetUint64(13285582464620353619)
	expected[8].SetUint64(11602188792716749495)
	expected[9].SetUint64(13293269979597702598)
	expected[10].SetUint64(17822114219392098785)
	expected[11].SetUint64(2946591587913066813)

	h := NewPermutation(12, 6, 17)
	h.Permutation(input[:])
	for i := 0; i < h.params.Width; i++ {
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mismatch error")
		}
	}
}

func BenchmarkPoseidon2Width8(b *testing.B) {
	h := NewPermutation(8, 6, 17)

	var tmp [8]fr.Element
	for i := range tmp {
		tmp[i].SetRandom()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Permutation(tmp[:])
	}
}

func BenchmarkPoseidon2Width12(b *testing.B) {
	h := NewPermutation(12, 6, 17)

	var tmp [12]fr.Element
	for i := range tmp {
		tmp[i].SetRandom()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Permutation(tmp[:])
	}
}
