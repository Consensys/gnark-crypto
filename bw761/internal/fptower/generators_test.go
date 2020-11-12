package fptower

import (
	"github.com/consensys/gurvy/bw761/fp"
	"github.com/consensys/gurvy/bw761/fr"
	"github.com/leanovate/gopter"
)

// TODO all gopter.Gen are incorrect, use same model as goff

// GenFp generates an Fp element
func GenFp() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var a0, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11 uint64
		a0 = genParams.NextUint64() % 17626244516597989515
		a1 = genParams.NextUint64() % 16614129118623039618
		a2 = genParams.NextUint64() % 1588918198704579639
		a3 = genParams.NextUint64() % 10998096788944562424
		a4 = genParams.NextUint64() % 8204665564953313070
		a5 = genParams.NextUint64() % 9694500593442880912
		a6 = genParams.NextUint64() % 274362232328168196
		a7 = genParams.NextUint64() % 8105254717682411801
		a8 = genParams.NextUint64() % 5945444129596489281
		a9 = genParams.NextUint64() % 13341377791855249032
		a10 = genParams.NextUint64() % 15098257552581525310
		a11 = genParams.NextUint64() % 81882988782276106
		elmt := fp.Element{
			a0, a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11,
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
		return &E2{A0: values[0].(fp.Element), A1: values[1].(fp.Element)}
	})
}

// GenE6 generates an E6 elmt
func GenE6() gopter.Gen {
	return gopter.CombineGens(
		GenE2(),
		GenE2(),
		GenE2(),
	).Map(func(values []interface{}) *E6 {
		return &E6{B0: *values[0].(*E2), B1: *values[1].(*E2), B2: *values[2].(*E2)}
	})
}

// ------------------------------------------------------------
// pairing generators

// GenFr generates an Fp element
func GenFr() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var a0, a1, a2, a3, a4, a5 uint64
		a0 = genParams.NextUint64() % 9586122913090633729
		a1 = genParams.NextUint64() % 1660523435060625408
		a2 = genParams.NextUint64() % 2230234197602682880
		a3 = genParams.NextUint64() % 1883307231910630287
		a4 = genParams.NextUint64() % 14284016967150029115
		a5 = genParams.NextUint64() % 121098312706494698
		elmt := fr.Element{
			a0, a1, a2, a3, a4, a5,
		}
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}
