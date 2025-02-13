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

	expected[0].SetUint64(715050894)
	expected[1].SetUint64(46053073)
	expected[2].SetUint64(1871489379)
	expected[3].SetUint64(80574331)
	expected[4].SetUint64(918832986)
	expected[5].SetUint64(249835165)
	expected[6].SetUint64(2075271471)
	expected[7].SetUint64(284356423)
	expected[8].SetUint64(1722012246)
	expected[9].SetUint64(1053014425)
	expected[10].SetUint64(747744298)
	expected[11].SetUint64(1087535683)
	expected[12].SetUint64(1042120606)
	expected[13].SetUint64(373122785)
	expected[14].SetUint64(67852658)
	expected[15].SetUint64(407644043)

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

	expected[0].SetUint64(1167703879)
	expected[1].SetUint64(298104798)
	expected[2].SetUint64(1345209527)
	expected[3].SetUint64(81383207)
	expected[4].SetUint64(1387678226)
	expected[5].SetUint64(518079145)
	expected[6].SetUint64(1565183874)
	expected[7].SetUint64(301357554)
	expected[8].SetUint64(869441538)
	expected[9].SetUint64(2130548890)
	expected[10].SetUint64(1046947186)
	expected[11].SetUint64(1913827299)
	expected[12].SetUint64(1269945069)
	expected[13].SetUint64(400345988)
	expected[14].SetUint64(1447450717)
	expected[15].SetUint64(183624397)
	expected[16].SetUint64(1181407857)
	expected[17].SetUint64(311808776)
	expected[18].SetUint64(1358913505)
	expected[19].SetUint64(95087185)
	expected[20].SetUint64(249194062)
	expected[21].SetUint64(1510301414)
	expected[22].SetUint64(426699710)
	expected[23].SetUint64(1293579823)

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
