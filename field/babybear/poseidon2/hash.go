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
	diag16[0].SetUint64(2013265919)
	diag16[1].SetUint64(1)
	diag16[2].SetUint64(2)
	diag16[3].SetUint64(1006632961)
	diag16[4].SetUint64(3)
	diag16[5].SetUint64(4)
	diag16[6].SetUint64(1006632960)
	diag16[7].SetUint64(2013265918)
	diag16[8].SetUint64(2013265917)
	diag16[9].SetUint64(2005401601)
	diag16[10].SetUint64(1509949441)
	diag16[11].SetUint64(1761607681)
	diag16[12].SetUint64(2013265906)
	diag16[13].SetUint64(7864320)
	diag16[14].SetUint64(125829120)
	diag16[15].SetUint64(15)

	// diagonal of internal matrix when Width=24
	diag24[0].SetUint64(2013265919)
	diag24[1].SetUint64(1)
	diag24[2].SetUint64(2)
	diag24[3].SetUint64(1006632961)
	diag24[4].SetUint64(3)
	diag24[5].SetUint64(4)
	diag24[6].SetUint64(1006632960)
	diag24[7].SetUint64(2013265918)
	diag24[8].SetUint64(2013265917)
	diag24[9].SetUint64(2005401601)
	diag24[10].SetUint64(1509949441)
	diag24[11].SetUint64(1761607681)
	diag24[12].SetUint64(1887436801)
	diag24[13].SetUint64(1997537281)
	diag24[14].SetUint64(2009333761)
	diag24[15].SetUint64(2013265906)
	diag24[16].SetUint64(7864320)
	diag24[17].SetUint64(503316480)
	diag24[18].SetUint64(251658240)
	diag24[19].SetUint64(125829120)
	diag24[20].SetUint64(62914560)
	diag24[21].SetUint64(31457280)
	diag24[22].SetUint64(15728640)
	diag24[23].SetUint64(15)
}
