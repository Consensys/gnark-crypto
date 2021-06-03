package bw6761

import (
	"crypto/rand"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bw6-761/fp"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/internal/fptower"
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

// GenE3 generates an E3 elmt
func GenE3() gopter.Gen {
	return gopter.CombineGens(
		GenFp(),
		GenFp(),
		GenFp(),
	).Map(func(values []interface{}) *fptower.E3 {
		return &fptower.E3{A0: values[0].(fp.Element), A1: values[1].(fp.Element), A2: values[2].(fp.Element)}
	})
}

// E6 generates an E6 elmt
func GenE6() gopter.Gen {
	return gopter.CombineGens(
		GenE3(),
		GenE3(),
	).Map(func(values []interface{}) *fptower.E6 {
		return &fptower.E6{B0: *values[0].(*fptower.E3), B1: *values[1].(*fptower.E3)}
	})
}

// GenBigInt generates a big.Int
func GenBigInt() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var s big.Int
		var b [fp.Bytes]byte
		_, err := rand.Read(b[:])
		if err != nil {
			panic(err)
		}
		s.SetBytes(b[:])
		genResult := gopter.NewGenResult(s, gopter.NoShrinker)
		return genResult
	}
}

// ------------------------------------------------------------
// pairing generators

// GenFr generates an Fp element
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
