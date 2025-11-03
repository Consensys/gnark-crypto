package hash_test

import (
	"bytes"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/poseidon2"
	"github.com/stretchr/testify/require"
)

func TestMerkleDamgardPadding(t *testing.T) {

	var (
		in          [fr.Bytes * 3]byte
		left, right []byte
		expectedSum []byte
		zero        [fr.Bytes]byte
	)
	for i := range len(in) / 2 {
		in[2*i] = byte(i / 256)
		in[2*i+1] = byte(i % 256)
	}

	const (
		pos1 = fr.Bytes
		pos2 = fr.Bytes + fr.Bytes/2 - 1
		pos3 = 2*fr.Bytes + fr.Bytes/2 + 5
		pos4 = 3 * fr.Bytes
	)

	permutation := poseidon2.NewDefaultPermutation()
	hash := poseidon2.NewMerkleDamgardHasher()

	// an entire element
	n, err := hash.Write(in[:pos1])
	require.NoError(t, err)
	require.Equal(t, pos1, n)
	sum := hash.Sum(nil)

	right = bytes.Clone(in[:pos1])
	left = make([]byte, fr.Bytes)
	expectedSum, err = permutation.Compress(left, right)
	require.NoError(t, err)

	require.Equal(t, expectedSum, sum)

	// a half-element
	left = expectedSum
	copy(right, zero[:])
	copy(right, in[pos1:pos2])
	expectedSum, err = permutation.Compress(left, right)
	require.NoError(t, err)

	n, err = hash.Write(in[pos1:pos2])
	require.NoError(t, err)
	require.Equal(t, pos2-pos1, n)
	sum = hash.Sum(nil)

	require.Equal(t, expectedSum, sum)

	// slightly larger than one element, overflows the buffer along the way
	left, err = permutation.Compress(left, in[fr.Bytes:2*fr.Bytes])
	copy(right, zero[:])
	copy(right, in[2*fr.Bytes:pos3])
	expectedSum, err = permutation.Compress(left, right)
	require.NoError(t, err)

	n, err = hash.Write(in[pos2:pos3])
	require.NoError(t, err)
	require.Equal(t, pos3-pos2, n)
	sum = hash.Sum(nil)

	require.Equal(t, expectedSum, sum)

	// less than half an element, stops after filling the buffer
	expectedSum, err = permutation.Compress(left, in[2*fr.Bytes:3*fr.Bytes])
	require.NoError(t, err)

	n, err = hash.Write(in[pos3:pos4])
	require.NoError(t, err)
	require.Equal(t, pos4-pos3, n)
	sum = hash.Sum(nil)

	require.Equal(t, expectedSum, sum)
}
