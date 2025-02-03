package poseidon2

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/hash"
)

// NewPoseidon2 returns a Poseidon2 hasher
// TODO @Tabaie @ThomasPiellard Generify once Poseidon2 parameters are known for all curves
func NewPoseidon2() hash.StateStorer {
	return hash.NewMerkleDamgardHasher(
		&Hash{params: parameters{
			t:         2,
			rF:        6,
			rP:        26,
			roundKeys: InitRC("Poseidon2 hash for BLS12-377 with t=2, rF=6, rP=26, d=17", 6, 26, 2),
		}}, make([]byte, fr.Bytes))
}
