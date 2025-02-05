package poseidon2

import (
	"hash"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
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
// - width: 2
// - nbFullRounds: 6
// - nbPartialRounds: 26
func NewDefaultParameters() *Parameters {
	return NewParameters(2, 6, 26)
}

func init() {
	gnarkHash.RegisterHash(gnarkHash.POSEIDON2_BLS12_377, func() hash.Hash {
		return NewMerkleDamgardHasher()
	})
}
