package vortex

import (
	"crypto/sha256"
	"hash"
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/stretchr/testify/require"
)

func TestPoseidon2BlockCompression(t *testing.T) {
	// This test ensures that the CompressPoseidon2 function is correctly implemented and produces the same output as
	// the poseidon2.NewMerkleDamgardHasher(), which uses Write and Sum methods to get the final hash output

	for i := 0; i < 100; i++ {
		var zero [8]koalabear.Element
		var input [8]koalabear.Element

		var inputBytes [32]byte
		for i := 0; i < 8; i++ {
			startIndex := i * 4
			input[i].SetRandom()
			valBytes := input[i].Bytes()
			copy(inputBytes[startIndex:startIndex+4], valBytes[:])
		}

		h := CompressPoseidon2(zero, input)

		merkleHasher := poseidon2.NewMerkleDamgardHasher()
		merkleHasher.Reset()
		merkleHasher.Write(inputBytes[:])
		newBytes := merkleHasher.Sum(nil)

		var result [8]koalabear.Element // Array to store the 8 reconstructed Elements

		for i := 0; i < 8; i++ {
			startIndex := i * 4
			segment := newBytes[startIndex : startIndex+4]
			var newElement koalabear.Element
			newElement.SetBytes(segment)
			result[i] = newElement
			require.Equal(t, result[i].String(), h[i].String())

		}

	}
}

func TestMerkleTree(t *testing.T) {

	posLists := []int{0, 1, 12, 31}

	t.Run("full-zero-leaves", func(t *testing.T) {
		assert := require.New(t)
		leaves := [32]Hash{}

		tree := BuildMerkleTree(leaves[:], nil)

		for _, pos := range posLists {

			proof, err := tree.Open(pos)
			assert.NoError(err)

			err = proof.Verify(pos, leaves[pos], tree.Root(), nil)
			assert.NoError(err)
		}
	})

	t.Run("full-random", func(t *testing.T) {
		assert := require.New(t)

		var (
			// #nosec G404 -- test case generation does not require a cryptographic PRNG
			rng     = rand.New(rand.NewChaCha8([32]byte{}))
			modulus = uint32(koalabear.Modulus().Int64())
		)

		leaves := [32]Hash{}
		for i := range leaves {
			for j := range leaves[i] {
				leaves[i][j] = koalabear.Element{rng.Uint32N(modulus)}
			}
		}

		tree := BuildMerkleTree(leaves[:], nil)

		for _, pos := range posLists {
			proof, err := tree.Open(pos)
			assert.NoError(err)

			err = proof.Verify(pos, leaves[pos], tree.Root(), nil)
			assert.NoError(err)
		}

	})

	t.Run("full-random-sha256", func(t *testing.T) {
		assert := require.New(t)

		var (
			// #nosec G404 -- test case generation does not require a cryptographic PRNG
			rng     = rand.New(rand.NewChaCha8([32]byte{}))
			modulus = uint32(koalabear.Modulus().Int64())
		)

		leaves := [32]Hash{}
		for i := range leaves {
			for j := range leaves[i] {
				leaves[i][j] = koalabear.Element{rng.Uint32N(modulus)}
			}
		}

		nh := func() hash.Hash { return sha256.New() }

		tree := BuildMerkleTree(leaves[:], nh)

		for _, pos := range posLists {
			proof, err := tree.Open(pos)
			assert.NoError(err)

			err = proof.Verify(pos, leaves[pos], tree.Root(), nh)
			assert.NoError(err)
		}

	})

}
