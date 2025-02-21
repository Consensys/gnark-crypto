package poseidon2

import (
	fr "github.com/consensys/gnark-crypto/field/koalabear"
)

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
var diag24 []fr.Element = make([]fr.Element, 24)

func init() {
	// diagonal of internal matrix when Width=16
	diag16[0].SetUint64(2130706431)
	diag16[1].SetUint64(1)
	diag16[2].SetUint64(2)
	diag16[3].SetUint64(1065353217)
	diag16[4].SetUint64(3)
	diag16[5].SetUint64(4)
	diag16[6].SetUint64(1065353216)
	diag16[7].SetUint64(2130706430)
	diag16[8].SetUint64(2130706429)
	diag16[9].SetUint64(2122383361)
	diag16[10].SetUint64(1864368129)
	diag16[11].SetUint64(2130706306)
	diag16[12].SetUint64(8323072)
	diag16[13].SetUint64(266338304)
	diag16[14].SetUint64(133169152)
	diag16[15].SetUint64(127)

	// diagonal of internal matrix when Width=24
	diag24[0].SetUint64(2130706431)
	diag24[1].SetUint64(1)
	diag24[2].SetUint64(2)
	diag24[3].SetUint64(1065353217)
	diag24[4].SetUint64(3)
	diag24[5].SetUint64(4)
	diag24[6].SetUint64(1065353216)
	diag24[7].SetUint64(2130706430)
	diag24[8].SetUint64(2130706429)
	diag24[9].SetUint64(2122383361)
	diag24[10].SetUint64(1598029825)
	diag24[11].SetUint64(1864368129)
	diag24[12].SetUint64(1997537281)
	diag24[13].SetUint64(2064121857)
	diag24[14].SetUint64(2097414145)
	diag24[15].SetUint64(2130706306)
	diag24[16].SetUint64(8323072)
	diag24[17].SetUint64(266338304)
	diag24[18].SetUint64(133169152)
	diag24[19].SetUint64(66584576)
	diag24[20].SetUint64(33292288)
	diag24[21].SetUint64(16646144)
	diag24[22].SetUint64(4161536)
	diag24[23].SetUint64(127)

}
