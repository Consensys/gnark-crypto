// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package hash_to_curve

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
)

// G1Sgn0 is an algebraic substitute for the notion of sign in ordered fields.
// Namely, every non-zero quadratic residue in a finite field of characteristic =/= 2 has exactly two square roots, one of each sign.
//
// See: https://www.rfc-editor.org/rfc/rfc9380.html#name-the-sgn0-function
//
// The sign of an element is not obviously related to that of its Montgomery form
func G1Sgn0(z *fp.Element) uint64 {

	nonMont := z.Bits()

	// m == 1
	return nonMont[0] % 2

}

func G1NotZero(x *fp.Element) uint64 {

	return x[0] | x[1] | x[2] | x[3]

}
