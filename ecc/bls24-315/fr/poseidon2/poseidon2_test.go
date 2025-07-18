// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package poseidon2

import (
	"crypto/rand"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	"github.com/consensys/gnark-crypto/hash"
	"github.com/stretchr/testify/require"
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

const msbMask = 0xff >> (9 - (fr.Bits % 8)) // to make sure randomized buffers are smaller than the modulus

func TestHashSmall(t *testing.T) {
	// hash two elements using Merkle-Damgard
	var b [2][fr.Bytes]byte
	h := NewMerkleDamgardHasher()
	for i := range b {
		_, err := rand.Read(b[i][:])
		require.NoError(t, err)
		b[i][0] &= msbMask
		_, err = h.Write(b[i][:])
		require.NoError(t, err)
	}
	p := Permutation{GetDefaultParameters()}
	res, err := p.Compress(make([]byte, fr.Bytes), b[0][:])
	require.NoError(t, err)
	res, err = p.Compress(res, b[1][:])
	require.NoError(t, err)

	require.Equal(t, res, h.Sum(nil))
}

func TestHashReset(t *testing.T) {
	// hash a single element using Merkle-Damgard and a nonzero IV, twice
	var iv, b [fr.Bytes]byte
	iv[0] = 1
	_, err := rand.Read(b[:])
	require.NoError(t, err)
	b[0] &= msbMask
	p := Permutation{GetDefaultParameters()}
	h := hash.NewMerkleDamgardHasher(&p, iv[:])
	_, err = h.Write(b[:])
	require.NoError(t, err)
	res := h.Sum(nil)

	h.Reset()
	_, err = h.Write(b[:])
	require.NoError(t, err)

	require.Equal(t, res, h.Sum(nil))
}
