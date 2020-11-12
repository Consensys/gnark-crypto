package bls377

import (
	"github.com/consensys/gurvy/bls377/fp"
	"github.com/consensys/gurvy/bls377/fr"
	"github.com/consensys/gurvy/bls377/internal/fptower"
	"github.com/leanovate/gopter"
)

// GenFp generates an Fp element
func GenFp() gopter.Gen {
	return func(genParams *gopter.GenParameters) *gopter.GenResult {
		var a0, a1, a2, a3, a4, a5 uint64
		a0 = genParams.NextUint64() % 9586122913090633729
		a1 = genParams.NextUint64() % 1660523435060625408
		a2 = genParams.NextUint64() % 2230234197602682880
		a3 = genParams.NextUint64() % 1883307231910630287
		a4 = genParams.NextUint64() % 14284016967150029115
		a5 = genParams.NextUint64() % 121098312706494698
		elmt := fp.Element{
			a0, a1, a2, a3, a4, a5,
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
		a0 = genParams.NextUint64() % 725501752471715841
		a1 = genParams.NextUint64() % 6461107452199829505
		a2 = genParams.NextUint64() % 6968279316240510977
		a3 = genParams.NextUint64() % 1345280370688173398
		elmt := fr.Element{
			a0, a1, a2, a3,
		}
		genResult := gopter.NewGenResult(elmt, gopter.NoShrinker)
		return genResult
	}
}
