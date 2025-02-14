// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package poseidon2

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/babybear"
)

func TestPoseidon2Width16(t *testing.T) {
	var input, expected [16]fr.Element
	input[0].SetUint64(894848333)
	input[1].SetUint64(1437655012)
	input[2].SetUint64(1200606629)
	input[3].SetUint64(1690012884)
	input[4].SetUint64(71131202)
	input[5].SetUint64(1749206695)
	input[6].SetUint64(1717947831)
	input[7].SetUint64(120589055)
	input[8].SetUint64(19776022)
	input[9].SetUint64(42382981)
	input[10].SetUint64(1831865506)
	input[11].SetUint64(724844064)
	input[12].SetUint64(171220207)
	input[13].SetUint64(1299207443)
	input[14].SetUint64(227047920)
	input[15].SetUint64(1783754913)

	expected[0].SetUint64(844700463)
	expected[1].SetUint64(487744727)
	expected[2].SetUint64(1398916173)
	expected[3].SetUint64(26581195)
	expected[4].SetUint64(20225485)
	expected[5].SetUint64(1676535670)
	expected[6].SetUint64(574441195)
	expected[7].SetUint64(1215372138)
	expected[8].SetUint64(222335125)
	expected[9].SetUint64(1878645310)
	expected[10].SetUint64(776550835)
	expected[11].SetUint64(1417481778)
	expected[12].SetUint64(1849058630)
	expected[13].SetUint64(1492102894)
	expected[14].SetUint64(390008419)
	expected[15].SetUint64(1030939362)

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
	input[0].SetUint64(89488333)
	input[1].SetUint64(143755012)
	input[2].SetUint64(120006629)
	input[3].SetUint64(169012884)
	input[4].SetUint64(7113202)
	input[5].SetUint64(174906695)
	input[6].SetUint64(171747831)
	input[7].SetUint64(12059055)
	input[8].SetUint64(1977022)
	input[9].SetUint64(4238981)
	input[10].SetUint64(183865506)
	input[11].SetUint64(72444064)
	input[12].SetUint64(17120207)
	input[13].SetUint64(129207443)
	input[14].SetUint64(22747920)
	input[15].SetUint64(178754913)
	input[16].SetUint64(89448333)
	input[17].SetUint64(143655012)
	input[18].SetUint64(120606629)
	input[19].SetUint64(169012884)
	input[20].SetUint64(7111202)
	input[21].SetUint64(174206695)
	input[22].SetUint64(171947831)
	input[23].SetUint64(12089055)

	expected[0].SetUint64(1305447350)
	expected[1].SetUint64(675392180)
	expected[2].SetUint64(675835241)
	expected[3].SetUint64(1994335439)
	expected[4].SetUint64(759610375)
	expected[5].SetUint64(129555205)
	expected[6].SetUint64(129998266)
	expected[7].SetUint64(1448498464)
	expected[8].SetUint64(1541233419)
	expected[9].SetUint64(911178249)
	expected[10].SetUint64(911621310)
	expected[11].SetUint64(216855587)
	expected[12].SetUint64(1806802708)
	expected[13].SetUint64(1176747538)
	expected[14].SetUint64(1177190599)
	expected[15].SetUint64(482424876)
	expected[16].SetUint64(1797470401)
	expected[17].SetUint64(1167415231)
	expected[18].SetUint64(1167858292)
	expected[19].SetUint64(473092569)
	expected[20].SetUint64(805435295)
	expected[21].SetUint64(175380125)
	expected[22].SetUint64(175823186)
	expected[23].SetUint64(1494323384)

	h := NewPermutation(24, 6, 19)
	h.Permutation(input[:])
	for i := 0; i < h.params.Width; i++ {
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mismatch error")
		}
	}
}

// bench
func BenchmarkPoseidon2Width16(b *testing.B) {
	h := NewPermutation(16, 6, 12)
	var tmp [16]fr.Element
	for i := 0; i < 16; i++ {
		tmp[i].SetRandom()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Permutation(tmp[:])
	}
}

func BenchmarkPoseidon2Width24(b *testing.B) {
	h := NewPermutation(24, 6, 19)
	var tmp [24]fr.Element
	for i := 0; i < 24; i++ {
		tmp[i].SetRandom()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Permutation(tmp[:])
	}
}
