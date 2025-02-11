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
	diag16[0].SetString("1526245758")
	diag16[1].SetString("1181039355")
	diag16[2].SetString("294525134")
	diag16[3].SetString("1187117748")
	diag16[4].SetString("1982584076")
	diag16[5].SetString("1805711612")
	diag16[6].SetString("1185627262")
	diag16[7].SetString("856598545")
	diag16[8].SetString("1032717003")
	diag16[9].SetString("1019193515")
	diag16[10].SetString("1622854951")
	diag16[11].SetString("1458848858")
	diag16[12].SetString("145265389")
	diag16[13].SetString("1220122282")
	diag16[14].SetString("1647429377")
	diag16[15].SetString("462996148")

	// diagonal of internal matrix when Width=24
	diag24[0].SetString("1151545613")
	diag24[1].SetString("1378328160")
	diag24[2].SetString("304145285")
	diag24[3].SetString("1170753223")
	diag24[4].SetString("1173904120")
	diag24[5].SetString("382806505")
	diag24[6].SetString("1018843483")
	diag24[7].SetString("326347087")
	diag24[8].SetString("1688249811")
	diag24[9].SetString("92343735")
	diag24[10].SetString("1989067774")
	diag24[11].SetString("426627352")
	diag24[12].SetString("1043427554")
	diag24[13].SetString("1350803455")
	diag24[14].SetString("685360948")
	diag24[15].SetString("992522095")
	diag24[16].SetString("461201261")
	diag24[17].SetString("546901022")
	diag24[18].SetString("1924506406")
	diag24[19].SetString("1858498981")
	diag24[20].SetString("227302405")
	diag24[21].SetString("642805551")
	diag24[22].SetString("522204904")
	diag24[23].SetString("1738963060")
}
