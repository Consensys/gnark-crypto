package hash_test

import (
	"crypto/rand"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon2"
	"github.com/stretchr/testify/require"
)

func TestMerkleDamgardHasherSum(t *testing.T) {
	const bn254MsbMask = 0xff >> (9 - fr.Bits%8)

	h := poseidon2.NewMerkleDamgardHasher()
	var b [2 * fr.Bytes]byte

	_, err := rand.Read(b[:])
	require.NoError(t, err)
	b[0] &= bn254MsbMask
	b[len(b)/2] &= bn254MsbMask

	_, err = h.Write(b[:])
	require.NoError(t, err)

	firstSum := h.Sum(nil)

	_, err = rand.Read(b[:])
	require.NoError(t, err)
	b[0] &= bn254MsbMask
	b[len(b)/2] &= bn254MsbMask

	secondSum := h.Sum(b[:])

	require.Equal(t, firstSum, h.State())

	_, err = h.Write(b[:])
	require.NoError(t, err)

	require.Equal(t, secondSum, h.Sum(nil))
}
