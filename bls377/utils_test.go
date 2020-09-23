package bls377

import (
	"github.com/consensys/gurvy/bls377/fp"
	"github.com/consensys/gurvy/bls377/fr"
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

// GenE2 generates an e2 elmt
func GenE2() gopter.Gen {
	return gopter.CombineGens(
		GenFp(),
		GenFp(),
	).Map(func(values []interface{}) *e2 {
		return &e2{values[0].(fp.Element), values[1].(fp.Element)}
	})
}

// GenE6 generates an e6 elmt
func GenE6() gopter.Gen {
	return gopter.CombineGens(
		GenE2(),
		GenE2(),
		GenE2(),
	).Map(func(values []interface{}) *e6 {
		return &e6{*values[0].(*e2), *values[1].(*e2), *values[2].(*e2)}
	})
}

// GenE12 generates an e6 elmt
func GenE12() gopter.Gen {
	return gopter.CombineGens(
		GenE6(),
		GenE6(),
	).Map(func(values []interface{}) *e12 {
		return &e12{*values[0].(*e6), *values[1].(*e6)}
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
