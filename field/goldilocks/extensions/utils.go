// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package extensions

import (
	"math/big"
	"sync"

	fr "github.com/consensys/gnark-crypto/field/goldilocks"
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
