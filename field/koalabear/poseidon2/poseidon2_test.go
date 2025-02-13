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

	expected[0].SetUint64(1068619505)
	expected[1].SetUint64(351695634)
	expected[2].SetUint64(60433440)
	expected[3].SetUint64(528072817)
	expected[4].SetUint64(590262369)
	expected[5].SetUint64(2004044931)
	expected[6].SetUint64(1712782737)
	expected[7].SetUint64(49715681)
	expected[8].SetUint64(580916736)
	expected[9].SetUint64(1994699298)
	expected[10].SetUint64(1703437104)
	expected[11].SetUint64(40370048)
	expected[12].SetUint64(2025966092)
	expected[13].SetUint64(1309042221)
	expected[14].SetUint64(1017780027)
	expected[15].SetUint64(1485419404)

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

	expected[0].SetUint64(1694135297)
	expected[1].SetUint64(2025543907)
	expected[2].SetUint64(964978921)
	expected[3].SetUint64(1324041389)
	expected[4].SetUint64(179393796)
	expected[5].SetUint64(510802406)
	expected[6].SetUint64(1580943853)
	expected[7].SetUint64(1940006321)
	expected[8].SetUint64(1206484229)
	expected[9].SetUint64(1537892839)
	expected[10].SetUint64(477327853)
	expected[11].SetUint64(836390321)
	expected[12].SetUint64(25634930)
	expected[13].SetUint64(357043540)
	expected[14].SetUint64(1427184987)
	expected[15].SetUint64(1786247455)
	expected[16].SetUint64(883104312)
	expected[17].SetUint64(1214512922)
	expected[18].SetUint64(153947936)
	expected[19].SetUint64(513010404)
	expected[20].SetUint64(1944813172)
	expected[21].SetUint64(145515349)
	expected[22].SetUint64(1215656796)
	expected[23].SetUint64(1574719264)

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
