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
// The default parameters are,
//
//  1. for compression:
//     - width: 16
//     - nbFullRounds: 6
//     - nbPartialRounds: 21
//
//  2. for sponge:
//     - width: 24
//     - nbFullRounds: 6
//     - nbPartialRounds: 21
func NewDefaultParameters() *Parameters {
	return NewParameters(16, 6, 21)
}

var diag16 [16]fr.Element
var diag24 [24]fr.Element

func init() {
	gnarkHash.RegisterHash(gnarkHash.POSEIDON2_KOALABEAR, func() hash.Hash {
		return NewMerkleDamgardHasher()
	})

	// diagonal of internal matrix when Width=16
	diag16[0].SetString("2046570709")
	diag16[1].SetString("758836515")
	diag16[2].SetString("1294135")
	diag16[3].SetString("1937509553")
	diag16[4].SetString("2128865551")
	diag16[5].SetString("1697674041")
	diag16[6].SetString("1520666926")
	diag16[7].SetString("1558715341")
	diag16[8].SetString("1137246046")
	diag16[9].SetString("1320163102")
	diag16[10].SetString("1968225990")
	diag16[11].SetString("234409115")
	diag16[12].SetString("946626146")
	diag16[13].SetString("566947014")
	diag16[14].SetString("2115407367")
	diag16[15].SetString("1174021429")

	// diagonal of internal matrix when Width=24
	diag24[0].SetString("1308809838")
	diag24[1].SetString("928239151")
	diag24[2].SetString("495882907")
	diag24[3].SetString("593757554")
	diag24[4].SetString("593757554")
	diag24[5].SetString("567559762")
	diag24[6].SetString("1572388064")
	diag24[7].SetString("1076816199")
	diag24[8].SetString("652906069")
	diag24[9].SetString("1871203714")
	diag24[10].SetString("358820701")
	diag24[11].SetString("1696335905")
	diag24[12].SetString("1771481038")
	diag24[13].SetString("67413782")
	diag24[14].SetString("1765474848")
	diag24[15].SetString("235952237")
	diag24[16].SetString("111716804")
	diag24[17].SetString("213766759")
	diag24[18].SetString("656473656")
	diag24[19].SetString("1879962596")
	diag24[20].SetString("1762516106")
	diag24[21].SetString("197180297")
	diag24[22].SetString("2061316832")
	diag24[23].SetString("1833346861")
}
