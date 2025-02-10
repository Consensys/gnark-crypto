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
// - width: 16
// - nbFullRounds: 6
// - nbPartialRounds: 12
func NewDefaultParameters() *Parameters {
	var diagonal [16]fr.Element
	diagonal[0].SetString("1337680893")
	diagonal[1].SetString("1849524340")
	diagonal[2].SetString("945441141")
	diagonal[3].SetString("110332385")
	diagonal[4].SetString("981849927")
	diagonal[5].SetString("511933108")
	diagonal[6].SetString("1289844587")
	diagonal[7].SetString("896077849")
	diagonal[8].SetString("971707481")
	diagonal[9].SetString("121792757")
	diagonal[10].SetString("1598170707")
	diagonal[11].SetString("1688648703")
	diagonal[12].SetString("26932898")
	diagonal[13].SetString("1760625654")
	diagonal[14].SetString("1480423439")
	diagonal[15].SetString("1903270600")
	return NewParameters(16, 6, 21, diagonal)
}

func init() {
	gnarkHash.RegisterHash(gnarkHash.POSEIDON2_BABYBEAR, func() hash.Hash {
		return NewMerkleDamgardHasher()
	})
}
