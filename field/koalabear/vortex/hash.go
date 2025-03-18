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
	x := append(a[:], b[:]...)
	if err := compressPerm.Permutation(x); err != nil {
		panic(err)
	}
	copy(res[:], x)
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

func HashPoseidon2x16(_x [][sisKeySize]koalabear.Element, merkleLeaves []Hash) {
	spongePerm.Permutation16x24(_x, merkleLeaves)
	// const (
	// 	width       = 16
	// 	p2blockSize = 16
	// 	stateSize   = 24
	// )
	// if len(_x) != width || len(merkleLeaves) != width {
	// 	panic("invalid input size")
	// }

	// var state [width][stateSize]koalabear.Element
	// n := len(_x[0])
	// m := len(Hash{})
	// for i := 0; i < n; i += p2blockSize {
	// 	for j := 0; j < width; j++ {
	// 		copy(state[j][m:], _x[j][i:])
	// 	}
	// 	spongePerm.Permutation16x24(&state)
	// 	// spongePerm.Permutation(state[j][:])
	// }

	// for i := range merkleLeaves {
	// 	copy(merkleLeaves[i][:], state[i][:])
	// }

	// for i, x := range _x {
	// 	var (
	// 		res Hash
	// 	)
	// 	for i := 0; i < len(x); i += p2blockSize {
	// 		copy(state[len(res):], x[i:])
	// 		spongePerm.Permutation(state[:])
	// 	}

	// 	copy(res[:], state[:])
	// 	merkleLeaves[i] = res
	// }
}
