package fptower

import (
	"crypto/rand"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
	"github.com/leanovate/gopter"
)

// Fp generates an Fp element
func GenFp() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var elmt fp.Element
		var b [fp.Bytes]byte
		rand.Read(b[:])
		elmt.SetBytes(b[:])
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}

// E2 generates an E2 elmt
func GenE2() gopter.Gen {
	return gopter.CombineGens(
		GenFp(),
		GenFp(),
	).Map(func(values []interface{}) *E2 {
		return &E2{A0: values[0].(fp.Element), A1: values[1].(fp.Element)}
	})
}

// E4 generates an E4 elmt
func GenE4() gopter.Gen {
	return gopter.CombineGens(
		GenE2(),
		GenE2(),
	).Map(func(values []interface{}) *E4 {
		return &E4{B0: *values[0].(*E2), B1: *values[1].(*E2)}
	})
}

// E8 generates an E8 elmt
func GenE8() gopter.Gen {
	return gopter.CombineGens(
		GenE4(),
		GenE4(),
	).Map(func(values []interface{}) *E8 {
		return &E8{C0: *values[0].(*E4), C1: *values[1].(*E4)}
	})
}

// E24 generates an E24 elmt
func GenE24() gopter.Gen {
	return gopter.CombineGens(
		GenE8(),
		GenE8(),
		GenE8(),
	).Map(func(values []interface{}) *E24 {
		return &E24{D0: *values[0].(*E8), D1: *values[1].(*E8), D2: *values[2].(*E8)}
	})
}
