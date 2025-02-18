package vortex

import (
	"testing"
)

func TestMerkleTree(t *testing.T) {

	t.Run("full-zero-leaves", func(t *testing.T) {

		leaves := [32]Hash{}

		tree := MerkleCompute(leaves[:])
		proof, err := tree.Open(0)
		if err != nil {
			t.Fatal(err)
		}

		if err := proof.Verify(0, leaves[0], tree.Root()); err != nil {
			t.Fatal(err)
		}
	})

}
