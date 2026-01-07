// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package secp256r1 efficient elliptic curve implementation for the NIST SECP56R1 (or P-256) curve (https://std.neuromancer.sk/secg/secp256r1).
//
// secp256r1: A j!=0 curve with
//
//		𝔽r: r=115792089210356248762697446949407573529996955224135760342422259061068512044369
//		𝔽p: p=115792089210356248762697446949407573530086143415290314195533631308867097853951
//		(E/𝔽p): Y²=X³+a*x+b where
//	     a=-3
//	     b=41058363725152142129326129780047268409114441015993725554835256314039467401291
//
// Security: estimated 128-bit level using Pollard's \rho attack
// (r is 256 bits)
//
// # Warning
//
// This code has been partially audited and is provided as-is. In particular, there is no security guarantees such as constant time implementation or side-channel attack resistance.
package secp256r1

import (
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/secp256r1/fp"
)

// ID secp256r1 ID
const ID = ecc.SECP256R1

// aCurveCoeff is the a coefficients of the curve Y²=X³+ax+b
var aCurveCoeff fp.Element
var bCurveCoeff fp.Element

// generator of the r-torsion group
var g1Gen G1Jac

var g1GenAff G1Affine

// point at infinity
var g1Infinity G1Jac

func init() {
	aCurveCoeff.SetInt64(-3)
	bCurveCoeff.SetString("41058363725152142129326129780047268409114441015993725554835256314039467401291")

	g1Gen.X.SetString("48439561293906451759052585252797914202762949526041747995844080717082404635286")
	g1Gen.Y.SetString("36134250956749795798585127919587881956611106672985015071877198253568414405109")
	g1Gen.Z.SetOne()

	g1GenAff.FromJacobian(&g1Gen)

	// (X,Y,Z) = (1,1,0)
	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()

}

// Generators return the generators of the r-torsion group, resp. in ker(pi-id), ker(Tr)
func Generators() (g1Jac G1Jac, g1Aff G1Affine) {
	g1Aff = g1GenAff
	g1Jac = g1Gen
	return
}

// CurveCoefficients returns the a, b coefficients of the curve equation.
func CurveCoefficients() (a, b fp.Element) {
	return aCurveCoeff, bCurveCoeff
}
