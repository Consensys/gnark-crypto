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

	// Create a buffer to hold the feed-forward input.
	copy(res[:], x[8:])
	if err := compressPerm.Permutation(x[:]); err != nil {
		// can't error (size is correct)
		panic(err)
	}

	for i := range res {
		res[i].Add(&res[i], &x[8+i])
	}
	return res
}

func CompressPoseidon2x16(matrix []koalabear.Element, colSize int, result []Hash) {
	compressPerm.Compressx16(matrix, colSize, result)
	// // ensure matrix has correct size
	// if len(matrix) != 16*colSize {
	// 	panic("invalid input size")
	// }
	// if len(result) != 16 {
	// 	panic("invalid output size")
	// }
	// if colSize%8 != 0 {
	// 	panic("invalid colSize, must be multiple of 8")
	// }
	// var x [16][16]koalabear.Element
	// nbSteps := colSize / 8
	// for step := 0; step < nbSteps; step++ {
	// 	// load chunk
	// 	for i := 0; i < 16; i++ {
	// 		// init state
	// 		copy(x[i][8:], matrix[i*colSize+step*8:i*colSize+step*8+8])
	// 		compressPerm.Permutation(x[i][:])
	// 		for j := 0; j < 8; j++ {
	// 			x[i][j].Add(&x[i][8+j], &matrix[i*colSize+step*8+j])
	// 		}
	// 	}
	// }
	// // store result
	// for i := 0; i < 16; i++ {
	// 	copy(result[i][:], x[i][:8])
	// }

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

func HashPoseidon2x16(sisHashes []koalabear.Element, merkleLeaves []Hash, sisKeySize int) {
	const (
		width       = 16
		p2blockSize = 16
		stateSize   = 24
	)
	if len(merkleLeaves) != width {
		panic("invalid input size")
	}

	var state [stateSize][width]koalabear.Element
	for i := 0; i < sisKeySize; i += p2blockSize {
		// transpose state
		for k := 8; k < stateSize; k++ {
			for j := 0; j < width; j++ {
				state[k][j] = sisHashes[j*sisKeySize+(k-8)+i]
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
