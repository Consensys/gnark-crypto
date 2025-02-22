package poseidon2

import (
	fr "github.com/consensys/gnark-crypto/field/goldilocks"
)

// NewParameters returns a new set of parameters for the Poseidon2 permutation.
// The default parameters are,
//
// - width: 8
// - nbFullRounds: 6
// - nbPartialRounds: 17
func NewDefaultParameters() *Parameters {
	return NewParameters(8, 6, 17)
}

var diag8 [8]fr.Element

func init() {
	// diagonal of internal matrix when Width=8
	// same as https://github.com/Plonky3/Plonky3/blob/f91c76545cf5c4ae9182897bcc557715817bcbdc/goldilocks/src/poseidon2.rs#L54
	diag8[0].SetUint64(0xa98811a1fed4e3a5)
	diag8[1].SetUint64(0x1cc48b54f377e2a0)
	diag8[2].SetUint64(0xe40cd4f6c5609a26)
	diag8[3].SetUint64(0x11de79ebca97a4a3)
	diag8[4].SetUint64(0x9177c73d8b7e929c)
	diag8[5].SetUint64(0x2a6fe8085797e791)
	diag8[6].SetUint64(0x3de6e93329f8d5ad)
	diag8[7].SetUint64(0x3f7af9125da962fe)
}
