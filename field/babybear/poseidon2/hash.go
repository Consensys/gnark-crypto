package poseidon2

import (
	"hash"

	fr "github.com/consensys/gnark-crypto/field/babybear"
	gnarkHash "github.com/consensys/gnark-crypto/hash"
)

// NewMerkleDamgardHasher returns a Poseidon2 hasher using the Merkle-Damgard
// construction with the default parameters.
func NewMerkleDamgardHasher() gnarkHash.StateStorer {
	// TODO @Tabaie @ThomasPiellard Generify once Poseidon2 parameters are known for all curves
	return gnarkHash.NewMerkleDamgardHasher(
		&Permutation{params: NewDefaultParameters()}, make([]byte, fr.Bytes))
}

// NewParameters returns a new set of parameters for the Poseidon2 permutation.
// The default parameters are:
// - width: 8
// - nbFullRounds: 6
// - nbPartialRounds: 10
func NewDefaultParameters() *Parameters {
	return NewParameters(8, 6, 10)
}

func init() {
	gnarkHash.RegisterHash(gnarkHash.POSEIDON2_BABYBEAR, func() hash.Hash {
		return NewMerkleDamgardHasher()
	})
}
