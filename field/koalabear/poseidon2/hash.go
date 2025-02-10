package poseidon2

import (
	"hash"

	fr "github.com/consensys/gnark-crypto/field/koalabear"
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
// - width: 16
// - nbFullRounds: 6
// - nbPartialRounds: 21
func NewDefaultParameters() *Parameters {
	var diagonal [16]fr.Element
	diagonal[0].SetString("2046570709")
	diagonal[1].SetString("758836515")
	diagonal[2].SetString("1294135")
	diagonal[3].SetString("1937509553")
	diagonal[4].SetString("2128865551")
	diagonal[5].SetString("1697674041")
	diagonal[6].SetString("1520666926")
	diagonal[7].SetString("1558715341")
	diagonal[8].SetString("1137246046")
	diagonal[9].SetString("1320163102")
	diagonal[10].SetString("1968225990")
	diagonal[11].SetString("234409115")
	diagonal[12].SetString("946626146")
	diagonal[13].SetString("566947014")
	diagonal[14].SetString("2115407367")
	diagonal[15].SetString("1174021429")
	return NewParameters(16, 6, 21, diagonal)
}

func init() {
	gnarkHash.RegisterHash(gnarkHash.POSEIDON2_KOALABEAR, func() hash.Hash {
		return NewMerkleDamgardHasher()
	})
}
