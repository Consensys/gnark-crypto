package bls12381

import (
	"math/rand"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/internal/fptower"
	"github.com/leanovate/gopter"
)

// ------------------------------------------------------------
// Tower generators

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

// GenE6 generates an fptower.E6 elmt
func GenE6() gopter.Gen {
	return gopter.CombineGens(
		GenE2(),
		GenE2(),
		GenE2(),
	).Map(func(values []interface{}) *fptower.E6 {
		return &fptower.E6{B0: *values[0].(*fptower.E2), B1: *values[1].(*fptower.E2), B2: *values[2].(*fptower.E2)}
	})
}

// GenE12 generates an fptower.E6 elmt
func GenE12() gopter.Gen {
	return gopter.CombineGens(
		GenE6(),
		GenE6(),
	).Map(func(values []interface{}) *fptower.E12 {
		return &fptower.E12{C0: *values[0].(*fptower.E6), C1: *values[1].(*fptower.E6)}
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
