// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package poseidon2

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/koalabear"
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

	expected[0].SetUint64(1716108683)
	expected[1].SetUint64(1764791125)
	expected[2].SetUint64(71140124)
	expected[3].SetUint64(832416356)
	expected[4].SetUint64(1404922729)
	expected[5].SetUint64(1453605171)
	expected[6].SetUint64(1890660603)
	expected[7].SetUint64(521230402)
	expected[8].SetUint64(862072475)
	expected[9].SetUint64(910754917)
	expected[10].SetUint64(1347810349)
	expected[11].SetUint64(2109086581)
	expected[12].SetUint64(140190770)
	expected[13].SetUint64(188873212)
	expected[14].SetUint64(625928644)
	expected[15].SetUint64(1387204876)

	h := NewPermutation(16, 6, 21)
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

	expected[0].SetUint64(334703116)
	expected[1].SetUint64(50967207)
	expected[2].SetUint64(1125931834)
	expected[3].SetUint64(2111316099)
	expected[4].SetUint64(1702309986)
	expected[5].SetUint64(1418574077)
	expected[6].SetUint64(362832271)
	expected[7].SetUint64(1348216536)
	expected[8].SetUint64(967163082)
	expected[9].SetUint64(683427173)
	expected[10].SetUint64(1758391800)
	expected[11].SetUint64(613069632)
	expected[12].SetUint64(854941900)
	expected[13].SetUint64(571205991)
	expected[14].SetUint64(1646170618)
	expected[15].SetUint64(500848450)
	expected[16].SetUint64(976613027)
	expected[17].SetUint64(692877118)
	expected[18].SetUint64(1767841745)
	expected[19].SetUint64(622519577)
	expected[20].SetUint64(1579142654)
	expected[21].SetUint64(1295406745)
	expected[22].SetUint64(239664939)
	expected[23].SetUint64(1225049204)

	h := NewPermutation(24, 6, 21)
	h.Permutation(input[:])
	for i := 0; i < h.params.Width; i++ {
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mismatch error")
		}
	}
}

// bench
func BenchmarkPoseidon2Width16(b *testing.B) {
	h := NewPermutation(16, 6, 21)
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
	h := NewPermutation(24, 6, 21)
	var tmp [24]fr.Element
	for i := 0; i < 24; i++ {
		tmp[i].SetRandom()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Permutation(tmp[:])
	}
}
