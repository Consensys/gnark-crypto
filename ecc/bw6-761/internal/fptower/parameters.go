// Copyright 2020 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bw6-761/fr"
)

// t-1
var xGen big.Int

var glvBasis ecc.Lattice

func init() {
	xGen.SetString("3362637538168598222219435186298528655381674028954528064283340709388076588006567983337308081752755143497537638367247", 10)
	_r := fr.Modulus()
	ecc.PrecomputeLattice(_r, &xGen, &glvBasis)
}
