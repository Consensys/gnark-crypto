package bn256

import (
	"github.com/consensys/gurvy/bn256/fp"
	"github.com/consensys/gurvy/bn256/fr"
	"github.com/leanovate/gopter"
)

// GenFp generates an Fp element
func GenFp() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var a0, a1, a2, a3 uint64
		a0 = genParams.NextUint64() % 4332616871279656263
		a1 = genParams.NextUint64() % 10917124144477883021
		a2 = genParams.NextUint64() % 13281191951274694749
		a3 = genParams.NextUint64() % 3486998266802970665
		elmt := fp.Element{
			a0, a1, a2, a3,
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
		a0 = genParams.NextUint64() % 4891460686036598785
		a1 = genParams.NextUint64() % 2896914383306846353
		a2 = genParams.NextUint64() % 13281191951274694749
		a3 = genParams.NextUint64() % 3486998266802970665
		elmt := fr.Element{
			a0, a1, a2, a3,
		}
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}
