package poseidon2

import (
	"hash"

	fr "github.com/consensys/gnark-crypto/field/babybear"
	gnarkHash "github.com/consensys/gnark-crypto/hash"
)

// NewMerkleDamgardHasher returns a Poseidon2 hasher using the Merkle-Damgard
// construction with the default parameters.
func NewMerkleDamgardHasher() gnarkHash.StateStorer {
	return gnarkHash.NewMerkleDamgardHasher(
		&Permutation{params: NewDefaultParameters()}, make([]byte, fr.Bytes))
}

// NewParameters returns a new set of parameters for the Poseidon2 permutation.
// The default parameters are,
//
//  1. for compression:
//     - width: 16
//     - nbFullRounds: 6
//     - nbPartialRounds: 12
//
//  2. for sponge:
//     - width: 24
//     - nbFullRounds: 6
//     - nbPartialRounds: 19
func NewDefaultParameters() *Parameters {
	return NewParameters(16, 6, 12)
	// return NewParameters(24, 6, 19)
}

var diag16 [16]fr.Element
var diag24 [24]fr.Element

func init() {
	gnarkHash.RegisterHash(gnarkHash.POSEIDON2_BABYBEAR, func() hash.Hash {
		return NewMerkleDamgardHasher()
	})

	// diagonal of internal matrix when Width=16
	diag16[0].SetString("2130706431")
	diag16[1].SetString("1")
	diag16[2].SetString("2")
	diag16[3].SetString("1065353217")
	diag16[4].SetString("3")
	diag16[5].SetString("4")
	diag16[6].SetString("1065353216")
	diag16[7].SetString("2130706430")
	diag16[8].SetString("2130706429")
	diag16[9].SetString("2122383361")
	diag16[10].SetString("1598029825")
	diag16[11].SetString("1864368129")
	diag16[12].SetString("1864368113")
	diag16[13].SetString("8323072")
	diag16[14].SetString("133169152")
	diag16[15].SetString("266338320")

	// diagonal of internal matrix when Width=24
	diag24[0].SetString("2130706431")
	diag24[1].SetString("1")
	diag24[2].SetString("2")
	diag24[3].SetString("1065353217")
	diag24[4].SetString("3")
	diag24[5].SetString("4")
	diag24[6].SetString("1065353216")
	diag24[7].SetString("2130706430")
	diag24[8].SetString("2130706429")
	diag24[9].SetString("2122383361")
	diag24[10].SetString("1598029825")
	diag24[11].SetString("1864368129")
	diag24[12].SetString("1997537281")
	diag24[13].SetString("2114060289")
	diag24[14].SetString("2126544897")
	diag24[15].SetString("1864368113")
	diag24[16].SetString("8323072")
	diag24[17].SetString("532676608")
	diag24[18].SetString("266338304")
	diag24[19].SetString("133169152")
	diag24[20].SetString("66584576")
	diag24[21].SetString("33292288")
	diag24[22].SetString("16646144")
	diag24[23].SetString("266338320")
}
