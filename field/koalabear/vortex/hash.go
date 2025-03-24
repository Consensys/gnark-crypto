package vortex

import (
	"github.com/consensys/gnark-crypto/field/koalabear"
	"github.com/consensys/gnark-crypto/field/koalabear/poseidon2"
)

var (
	// compressPerm stores the parameters of the poseidon2 permutation
	// that we use for merkle trees.
	compressPerm = poseidon2.NewPermutation(16, 6, 21)
	// spongePerm stores the parameters of the poseidon2 permutation
	// that we use for the sponge construction.
	spongePerm = poseidon2.NewPermutation(24, 6, 21)
)

// CompressPoseidon2 runs the Poseidon2 compression function over two hashes
func CompressPoseidon2(a, b Hash) Hash {
	res := Hash{}
	var x [16]koalabear.Element
	copy(x[:], a[:])
	copy(x[8:], b[:])
	if err := compressPerm.Permutation(x[:]); err != nil {
		panic(err)
	}
	copy(res[:], x[:8])
	return res
}

// HashPoseidon2 returns a Poseidon2 hash of an array of field elements. The
// input is zero-padded so it should be used only in the context of fixed
// length hashes to avoid padding attacks.
func HashPoseidon2(x []koalabear.Element) Hash {

	const (
		blockSize = 16
		stateSize = 24
	)
	var (
		res   Hash
		state [stateSize]koalabear.Element
	)

	for i := 0; i < len(x); i += blockSize {
		copy(state[len(res):], x[i:])
		spongePerm.Permutation(state[:])
	}

	copy(res[:], state[:])
	return res
}

func HashPoseidon2x16(_x []koalabear.Element, merkleLeaves []Hash) {
	const (
		width       = 16
		p2blockSize = 16
		stateSize   = 24
	)
	if len(_x)/sisKeySize != width || len(merkleLeaves) != width {
		panic("invalid input size")
	}

	var state [stateSize][width]koalabear.Element
	n := sisKeySize
	for i := 0; i < n; i += p2blockSize {
		// transpose state
		for k := 8; k < stateSize; k++ {
			for j := 0; j < width; j++ {
				state[k][j] = _x[j*n+(k-8)+i]
			}
		}
		spongePerm.Permutation16x24(&state)
	}

	// transpose back the first 8 into merkleLeaves
	for k := 0; k < 8; k++ {
		for j := 0; j < width; j++ {
			merkleLeaves[j][k] = state[k][j]
		}
	}
}
