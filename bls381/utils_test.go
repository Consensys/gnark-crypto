package bls381

import (
	"github.com/consensys/gurvy/bls381/fp"
	"github.com/consensys/gurvy/bls381/fr"
	"github.com/leanovate/gopter"
)

// ------------------------------------------------------------
// Tower generators

// GenFp generates an Fp element
func GenFp() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var a0, a1, a2, a3, a4, a5 uint64
		a0 = genParams.NextUint64() % 13402431016077863595
		a1 = genParams.NextUint64() % 2210141511517208575
		a2 = genParams.NextUint64() % 7435674573564081700
		a3 = genParams.NextUint64() % 7239337960414712511
		a4 = genParams.NextUint64() % 5412103778470702295
		a5 = genParams.NextUint64() % 1873798617647539866
		elmt := fp.Element{
			a0, a1, a2, a3, a4, a5,
		}
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}

// GenE2 generates an E2 elmt
func GenE2() gopter.Gen {
	return gopter.CombineGens(
		GenFp(),
		GenFp(),
	).Map(func(values []interface{}) *E2 {
		return &E2{values[0].(fp.Element), values[1].(fp.Element)}
	})
}

// GenE6 generates an E6 elmt
func GenE6() gopter.Gen {
	return gopter.CombineGens(
		GenE2(),
		GenE2(),
		GenE2(),
	).Map(func(values []interface{}) *E6 {
		return &E6{*values[0].(*E2), *values[1].(*E2), *values[2].(*E2)}
	})
}

// GenE12 generates an E6 elmt
func GenE12() gopter.Gen {
	return gopter.CombineGens(
		GenE6(),
		GenE6(),
	).Map(func(values []interface{}) *E12 {
		return &E12{*values[0].(*E6), *values[1].(*E6)}
	})
}

// ------------------------------------------------------------
// pairing generators

// GenFr generates an Fr element
func GenFr() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var a0, a1, a2, a3 uint64
		a0 = genParams.NextUint64() % 18446744069414584321
		a1 = genParams.NextUint64() % 6034159408538082302
		a2 = genParams.NextUint64() % 3691218898639771653
		a3 = genParams.NextUint64() % 8353516859464449352
		elmt := fr.Element{
			a0, a1, a2, a3,
		}
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}
