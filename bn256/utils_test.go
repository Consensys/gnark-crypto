package bn256

import (
	"github.com/consensys/gurvy/bn256/fp"
	"github.com/consensys/gurvy/bn256/fr"
	"github.com/consensys/gurvy/bn256/internal/fptower"
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
