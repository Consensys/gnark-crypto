// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package poseidon2

import (
	"testing"

	fr "github.com/consensys/gnark-crypto/field/goldilocks"
)

func TestPoseidon2Width8(t *testing.T) {
	var input, expected [8]fr.Element
	input[0].SetUint64(5116996373749832116)
	input[1].SetUint64(8931548647907683339)
	input[2].SetUint64(17132360229780760684)
	input[3].SetUint64(11280040044015983889)
	input[4].SetUint64(11957737519043010992)
	input[5].SetUint64(15695650327991256125)
	input[6].SetUint64(17604752143022812942)
	input[7].SetUint64(543194415197607509)

	expected[0].SetUint64(13660962709356710016)
	expected[1].SetUint64(14157778087516188574)
	expected[2].SetUint64(16152999951555461431)
	expected[3].SetUint64(8607075105484324444)
	expected[4].SetUint64(1743726080892756454)
	expected[5].SetUint64(2240541459052235012)
	expected[6].SetUint64(4235763323091507869)
	expected[7].SetUint64(15136582546434955203)

	h := NewPermutation(8, 6, 17)
	h.Permutation(input[:])
	for i := 0; i < h.params.Width; i++ {
		if !input[i].Equal(&expected[i]) {
			t.Fatal("mismatch error")
		}
	}
}

// bench
func BenchmarkPoseidon2Width8(b *testing.B) {
	h := NewPermutation(8, 6, 17)
	var tmp [8]fr.Element
	for i := 0; i < 8; i++ {
		tmp[i].SetRandom()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Permutation(tmp[:])
	}
}
