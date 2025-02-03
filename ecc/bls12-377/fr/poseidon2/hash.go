package poseidon2

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/consensys/gnark-crypto/hash"
)

func NewPoseidon2() hash.StateStorer {
	return hash.NewMerkleDamgardHasher(&Hash{params: GetParameters()}, make([]byte, fr.Bytes))
}

// TODO @Tabaie @ThomasPiellard Generify once Poseidon2 parameters are known for all curves
func GetParameters() Parameters {
	return Parameters{
		T:         2,
		Rf:        6,
		Rp:        26,
		RoundKeys: InitRC("Poseidon2 hash for BLS12-377 with t=2, rF=6, rP=26, d=17", 6, 26, 2),
	}
}
