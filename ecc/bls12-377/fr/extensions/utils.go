package extensions

import (
	"math/big"
	"sync"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/leanovate/gopter"
)

var bigIntPool = sync.Pool{
	New: func() interface{} {
		return new(big.Int)
	},
}

// Fr generates an Fr element
func GenFr() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var elmt fr.Element

		if _, err := elmt.SetRandom(); err != nil {
			panic(err)
		}
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}

// E2 generates an E2 elmt
func GenE2() gopter.Gen {
	return gopter.CombineGens(
		GenFr(),
		GenFr(),
	).Map(func(values []interface{}) *E2 {
		return &E2{A0: values[0].(fr.Element), A1: values[1].(fr.Element)}
	})
}

// GenE3 generates an E3 elmt
func GenE3() gopter.Gen {
	return gopter.CombineGens(
		GenFr(),
		GenFr(),
		GenFr(),
	).Map(func(values []interface{}) *E3 {
		return &E3{A0: values[0].(fr.Element), A1: values[1].(fr.Element), A2: values[2].(fr.Element)}
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
