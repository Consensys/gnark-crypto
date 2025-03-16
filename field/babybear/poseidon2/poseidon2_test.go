// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package poseidon2

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/babybear"
	"github.com/consensys/gnark-crypto/utils/cpu"
	"github.com/stretchr/testify/require"
)

func TestMulMulInternalInPlaceWidth16(t *testing.T) {
	var input, expected [16]fr.Element
	for i := range input {
		input[i].SetRandom()
	}

	expected = input

	h := NewPermutation(16, 8, 13)
	h.matMulInternalInPlace(expected[:])

	var sum fr.Element
	sum.Set(&input[0])
	for i := 1; i < h.params.Width; i++ {
		sum.Add(&sum, &input[i])
	}
	for i := 0; i < h.params.Width; i++ {
		input[i].Mul(&input[i], &diag16[i]).
			Add(&input[i], &sum)
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mat mul internal w/ diagonal doesn't match hand calculated")
		}
	}
}

func TestMulMulInternalInPlaceWidth24(t *testing.T) {
	var input, expected [24]fr.Element
	for i := range input {
		input[i].SetRandom()
	}

	expected = input

	h := NewPermutation(24, 8, 21)
	h.matMulInternalInPlace(expected[:])

	var sum fr.Element
	sum.Set(&input[0])
	for i := 1; i < h.params.Width; i++ {
		sum.Add(&sum, &input[i])
	}
	for i := 0; i < h.params.Width; i++ {
		input[i].Mul(&input[i], &diag24[i]).
			Add(&input[i], &sum)
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mat mul internal w/ diagonal doesn't match hand calculated")
		}
	}
}

func TestAVX512Width16(t *testing.T) {
	if !cpu.SupportAVX512 {
		t.Skip("AVX512 not supported")
	}
	assert := require.New(t)
	var input, expected [16]fr.Element
	for i := range input {
		input[i].SetRandom()
	}

	expected = input

	h := NewPermutation(16, 8, 13)

	err := h.Permutation(input[:])
	assert.NoError(err)

	h.disableAVX512()
	err = h.Permutation(expected[:])
	assert.NoError(err)

	// compare results
	for i := 0; i < h.params.Width; i++ {
		assert.True(input[i].Equal(&expected[i]), "avx512 result don't match purego")
	}
}

func TestAVX512Width24(t *testing.T) {
	if !cpu.SupportAVX512 {
		t.Skip("AVX512 not supported")
	}
	assert := require.New(t)
	var input, expected [24]fr.Element
	for i := range input {
		input[i].SetRandom()
	}

	expected = input

	h := NewPermutation(24, 8, 21)

	err := h.Permutation(input[:])
	assert.NoError(err)

	h.disableAVX512()
	err = h.Permutation(expected[:])
	assert.NoError(err)

	// compare results
	for i := 0; i < h.params.Width; i++ {
		assert.True(input[i].Equal(&expected[i]), "avx512 result don't match purego")
	}
}

func TestAVX512Permutation16x24(t *testing.T) {
	if !cpu.SupportAVX512 {
		t.Skip("AVX512 not supported")
	}
	assert := require.New(t)
	input := make([][512]fr.Element, 16)
	result := make([][8]fr.Element, 16)
	expected := make([][8]fr.Element, 16)

	for i := range input {
		for j := range input[i] {
			input[i][j].SetRandom()
		}
	}
	h := NewPermutation(24, 8, 21)

	h.Permutation16x24(input, result)

	h.disableAVX512()
	h.Permutation16x24(input, expected)

	// compare results
	for i := 0; i < 16; i++ {
		for j := 0; j < 8; j++ {
			assert.True(result[i][j].Equal(&expected[i][j]), "avx512 result don't match purego")
		}
	}
}

func BenchmarkPermutation16x24(b *testing.B) {
	input := make([][512]fr.Element, 16)
	res := make([][8]fr.Element, 16)
	for i := range input {
		for j := range input[i] {
			input[i][j].SetRandom()
		}
	}
	h := NewPermutation(24, 8, 21)

	b.SetBytes(16 * 512 * 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Permutation16x24(input, res)
	}
}

func (h *Permutation) disableAVX512() {
	h.params.hasFast16_8_13 = false
	h.params.hasFast24_8_21 = false
}

func TestPoseidon2Width16(t *testing.T) {
	var input, expected [16]fr.Element
	// these are random values generated by SetRandom()
	input[0].SetUint64(926848709)
	input[1].SetUint64(772257670)
	input[2].SetUint64(775357184)
	input[3].SetUint64(1501166730)
	input[4].SetUint64(865948535)
	input[5].SetUint64(1208358603)
	input[6].SetUint64(1755902432)
	input[7].SetUint64(392259314)
	input[8].SetUint64(630678817)
	input[9].SetUint64(1665029989)
	input[10].SetUint64(1776916052)
	input[11].SetUint64(36754593)
	input[12].SetUint64(1920998735)
	input[13].SetUint64(842665326)
	input[14].SetUint64(1674852701)
	input[15].SetUint64(310605518)

	expected[0].SetUint64(818741542)
	expected[1].SetUint64(742709230)
	expected[2].SetUint64(1128775763)
	expected[3].SetUint64(1028903280)
	expected[4].SetUint64(90185980)
	expected[5].SetUint64(263112871)
	expected[6].SetUint64(1128687407)
	expected[7].SetUint64(1726949704)
	expected[8].SetUint64(1079297148)
	expected[9].SetUint64(1309030355)
	expected[10].SetUint64(1596410868)
	expected[11].SetUint64(869945617)
	expected[12].SetUint64(1079234851)
	expected[13].SetUint64(1064884418)
	expected[14].SetUint64(1362602666)
	expected[15].SetUint64(652219983)

	h := NewPermutation(16, 6, 12)
	h.Permutation(input[:])
	for i := 0; i < h.params.Width; i++ {
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mismatch error")
		}
	}
}

func TestPoseidon2Width24(t *testing.T) {
	var input, expected [24]fr.Element
	// these are random values generated by SetRandom()
	input[0].SetUint64(60806399)
	input[1].SetUint64(523046893)
	input[2].SetUint64(770765907)
	input[3].SetUint64(316416977)
	input[4].SetUint64(214364663)
	input[5].SetUint64(1341870810)
	input[6].SetUint64(1556213068)
	input[7].SetUint64(175271367)
	input[8].SetUint64(1651721560)
	input[9].SetUint64(1496696610)
	input[10].SetUint64(1823989412)
	input[11].SetUint64(1045720388)
	input[12].SetUint64(1480044199)
	input[13].SetUint64(698921269)
	input[14].SetUint64(163319479)
	input[15].SetUint64(1553935046)
	input[16].SetUint64(1332517615)
	input[17].SetUint64(1026652696)
	input[18].SetUint64(1770706686)
	input[19].SetUint64(1656168728)
	input[20].SetUint64(1447871165)
	input[21].SetUint64(1397927099)
	input[22].SetUint64(641149593)
	input[23].SetUint64(1002972123)

	expected[0].SetUint64(1487985473)
	expected[1].SetUint64(854017561)
	expected[2].SetUint64(308629844)
	expected[3].SetUint64(1234724305)
	expected[4].SetUint64(741681298)
	expected[5].SetUint64(384142256)
	expected[6].SetUint64(1247322610)
	expected[7].SetUint64(323136600)
	expected[8].SetUint64(173214613)
	expected[9].SetUint64(144598085)
	expected[10].SetUint64(1033718386)
	expected[11].SetUint64(273587448)
	expected[12].SetUint64(2009882407)
	expected[13].SetUint64(1737843408)
	expected[14].SetUint64(1245051692)
	expected[15].SetUint64(1020306129)
	expected[16].SetUint64(486901205)
	expected[17].SetUint64(584997799)
	expected[18].SetUint64(1607945291)
	expected[19].SetUint64(1919345816)
	expected[20].SetUint64(1130841387)
	expected[21].SetUint64(56863906)
	expected[22].SetUint64(1336666656)
	expected[23].SetUint64(370881953)

	h := NewPermutation(24, 6, 19)
	h.Permutation(input[:])
	for i := 0; i < h.params.Width; i++ {
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mismatch error")
		}
	}
}

func BenchmarkPoseidon2Width16(b *testing.B) {
	h := NewPermutation(16, 8, 13)

	var tmp [16]fr.Element
	for i := range tmp {
		tmp[i].SetRandom()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Permutation(tmp[:])
	}
}

func BenchmarkPoseidon2Width24(b *testing.B) {
	h := NewPermutation(24, 8, 21)

	var tmp [24]fr.Element
	for i := range tmp {
		tmp[i].SetRandom()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Permutation(tmp[:])
	}
}
