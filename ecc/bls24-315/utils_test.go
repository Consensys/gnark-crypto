package bls24315

import (
	"math/rand"

	"github.com/consensys/gnark-crypto/ecc/bls24-315/fp"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/fr"
	"github.com/consensys/gnark-crypto/ecc/bls24-315/internal/fptower"
	"github.com/leanovate/gopter"
)

// GenFp generates an Fp element
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

// GenE2 generates an fptower.E2 elmt
func GenE2() gopter.Gen {
	return gopter.CombineGens(
		GenFp(),
		GenFp(),
	).Map(func(values []interface{}) *fptower.E2 {
		return &fptower.E2{A0: values[0].(fp.Element), A1: values[1].(fp.Element)}
	})
}

// GenE4 generates an fptower.E4 elmt
func GenE4() gopter.Gen {
	return gopter.CombineGens(
		GenE2(),
		GenE2(),
	).Map(func(values []interface{}) *fptower.E4 {
		return &fptower.E4{B0: *values[0].(*fptower.E2), B1: *values[1].(*fptower.E2)}
	})
}

// GenE8 generates an fptower.E8 elmt
func GenE8() gopter.Gen {
	return gopter.CombineGens(
		GenE4(),
		GenE4(),
	).Map(func(values []interface{}) *fptower.E8 {
		return &fptower.E8{C0: *values[0].(*fptower.E4), C1: *values[1].(*fptower.E4)}
	})
}

// GenE24 generates an fptower.E24 elmt
func GenE24() gopter.Gen {
	return gopter.CombineGens(
		GenE8(),
		GenE8(),
		GenE8(),
	).Map(func(values []interface{}) *fptower.E24 {
		return &fptower.E24{D0: *values[0].(*fptower.E8), D1: *values[1].(*fptower.E8), D2: *values[2].(*fptower.E8)}
	})
}

// ------------------------------------------------------------
// pairing generators

// GenFr generates an Fr element
func GenFr() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var elmt fr.Element
		var b [fr.Bytes]byte
		rand.Read(b[:])
		elmt.SetBytes(b[:])
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}
