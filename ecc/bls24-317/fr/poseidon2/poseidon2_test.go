// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package poseidon2

import (
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
)

func TestExternalMatrix(t *testing.T) {
	t.Skip("skipping test - it is initialized for width=4 for which we don't have the diagonal matrix")

	var expected [4][4]fr.Element
	expected[0][0].SetUint64(5)
	expected[0][1].SetUint64(4)
	expected[0][2].SetUint64(1)
	expected[0][3].SetUint64(1)

	expected[1][0].SetUint64(7)
	expected[1][1].SetUint64(6)
	expected[1][2].SetUint64(3)
	expected[1][3].SetUint64(1)

	expected[2][0].SetUint64(1)
	expected[2][1].SetUint64(1)
	expected[2][2].SetUint64(5)
	expected[2][3].SetUint64(4)

	expected[3][0].SetUint64(3)
	expected[3][1].SetUint64(1)
	expected[3][2].SetUint64(7)
	expected[3][3].SetUint64(6)

	h := NewPermutation(4, 8, 56)
	var tmp [4]fr.Element
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			tmp[j].SetUint64(0)
			if i == j {
				tmp[j].SetOne()
			}
		}
		// h.Write(tmp[:])
		h.matMulExternalInPlace(tmp[:])
		for j := 0; j < 4; j++ {
			if !tmp[j].Equal(&expected[i][j]) {
				t.Fatal("error matMul4")
			}
		}
	}

}

func BenchmarkPoseidon2(b *testing.B) {
	h := NewPermutation(3, 8, 56)
	var tmp [3]fr.Element
	fr.Vector(tmp[:]).MustSetRandom()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Permutation(tmp[:])
	}
}
