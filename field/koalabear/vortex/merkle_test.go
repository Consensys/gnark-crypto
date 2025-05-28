package vortex

import (
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/stretchr/testify/require"
)

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

}
