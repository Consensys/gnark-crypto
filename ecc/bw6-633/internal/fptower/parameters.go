// Copyright 2020-2024 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bw6-633/fr"
)

// t-1
var xGen big.Int

var glvBasis ecc.Lattice

func init() {
	xGen.SetString("37014442673353839783463348892746893664389658635873267609916377398480286678854893830143", 10)
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &xGen, &glvBasis)
}
