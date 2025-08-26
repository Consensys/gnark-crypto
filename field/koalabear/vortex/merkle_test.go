package vortex

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"math/rand/v2"
	"testing"

	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
	"github.com/stretchr/testify/require"
)

func TestMulMulInternal(t *testing.T) {
	// var input, expected [16]fr.Element
	var old, block Hash
	old[0] = koalabear.NewElement(uint64(703724752))
	old[1] = koalabear.NewElement(uint64(280040542))
	old[2] = koalabear.NewElement(uint64(1514240686))
	old[3] = koalabear.NewElement(uint64(986917665))
	old[4] = koalabear.NewElement(uint64(1972211392))
	old[5] = koalabear.NewElement(uint64(832875602))
	old[6] = koalabear.NewElement(uint64(2095324332))
	old[7] = koalabear.NewElement(uint64(36857942))

	block[0] = koalabear.NewElement(uint64(760417386))
	block[1] = koalabear.NewElement(uint64(1333026101))
	block[2] = koalabear.NewElement(uint64(835587083))
	block[3] = koalabear.NewElement(uint64(1017667263))
	block[4] = koalabear.NewElement(uint64(669624325))
	block[5] = koalabear.NewElement(uint64(1903375813))
	block[6] = koalabear.NewElement(uint64(1853215757))
	block[7] = koalabear.NewElement(uint64(199352308))
	expect := CompressPoseidon2(old, block)

	fmt.Printf("input=%v\n", old)
	fmt.Printf("expect=%v\n", expect)

	// fmt.Printf("expected=%v\n", expected)

}

func TestMulMulInternalInPlaceWidth16(t *testing.T) {
	var input [16]koalabear.Element
	for i := range input {
		input[i] = koalabear.NewElement(uint64(i))
	}

	var hash1, hash2 Hash
	copy(hash1[:], input[:8])
	copy(hash2[:], input[8:])
	expected := CompressPoseidon2(hash1, hash2)

	fmt.Printf("input=%v\n", input)

	fmt.Printf("expected=%v\n", expected)

}
func TestPoseidon2singleBlockCompression(t *testing.T) {
	// This test ensures that the CompressPoseidon2 function is correctly implemented and produces the same output as
	// the poseidon2.NewMerkleDamgardHasher(), which uses Write and Sum methods to get the final hash output

	var state [8]koalabear.Element
	var input [8]koalabear.Element

	merkleHasher := poseidon2.NewMerkleDamgardHasher()
	merkleHasher.Reset()

	for i := 0; i < 2; i++ {
		input[7].SetRandom()
		state = CompressPoseidon2(state, input)
		inputBytes := input[7].Bytes()
		merkleHasher.Write(inputBytes[:])

	}

	// h := poseidon2.BlockCompressionMekle(state, input)

	// merkleHasher.Write(inputBytes[:])
	newBytes := merkleHasher.Sum(nil)

	var result [8]koalabear.Element // Array to store the 8 reconstructed Elements

	for i := 0; i < 8; i++ {
		startIndex := i * 4
		segment := newBytes[startIndex : startIndex+4]
		var newElement koalabear.Element
		newElement.SetBytes(segment)
		result[i] = newElement
		require.Equal(t, result[i].String(), state[i].String())

	}

}

func TestPoseidon2BlockCompression(t *testing.T) {
	// This test ensures that the Poseidon2BlockCompression function is correctly implemented and produces the same output as
	// the poseidon2.NewMerkleDamgardHasher(), which uses Write and Sum methods to get the final hash output

	// We hash and compress one Octuplet at a time

	for i := 0; i < 1; i++ {
		var state [8]koalabear.Element
		var input [8]koalabear.Element

		inputBytes := make([]byte, 9*4)

		for i := 0; i < 8; i++ {
			startIndex := i * 4
			input[i].SetRandom()
			valBytes := input[i].Bytes()
			copy(inputBytes[startIndex:startIndex+4], valBytes[:])
		}
		var val koalabear.Element
		val.SetRandom()
		valBytes := val.Bytes()
		copy(inputBytes[32:36], valBytes[:])

		h := CompressPoseidon2(state, input)

		merkleHasher := poseidon2.NewMerkleDamgardHasher()
		merkleHasher.Reset()
		merkleHasher.Write(inputBytes[:32]) // write one 32 bytes (equivalent to one Octuplet)
		newBytes := merkleHasher.Sum(nil)

		var result [8]koalabear.Element

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

func TestPoseidon2SpongeCompression(t *testing.T) {
	// This test ensures that the CompressPoseidon2 function is correctly implemented and produces the same output as
	// the poseidon2.NewMerkleDamgardHasher(), which uses Write and Sum methods to get the final hash output

	var input [8]koalabear.Element

	var inputBytes [32]byte
	for i := 0; i < 8; i++ {
		startIndex := i * 4
		// input[i].SetRandom()
		valBytes := input[i].Bytes()
		copy(inputBytes[startIndex:startIndex+4], valBytes[:])
	}

	h := HashPoseidon2(input[:])

	merkleHasher := poseidon2.NewMerkleDamgardHasher24()
	merkleHasher.Reset()
	merkleHasher.Write(inputBytes[:])
	newBytes := merkleHasher.Sum(nil)

	fmt.Printf("newBytes %v\n", len(newBytes))
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
