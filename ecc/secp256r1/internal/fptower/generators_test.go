package fp2

import (
	"github.com/consensys/gnark-crypto/ecc/secp256r1/fp"
	"github.com/leanovate/gopter"
)

func GenFp() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var elmt fp.Element
		elmt.MustSetRandom()
		return gopter.NewGenResult(elmt, gopter.NoShrinker)
	}
}

func GenE2() gopter.Gen {
	return gopter.CombineGens(
		GenFp(),
		GenFp(),
	).Map(func(values []any) *E2 {
		return &E2{A0: values[0].(fp.Element), A1: values[1].(fp.Element)}
	})
}
