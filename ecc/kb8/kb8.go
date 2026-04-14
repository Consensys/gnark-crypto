// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

// Package kb8 for efficient elliptic curve implementation for kb8 (koalabear-8).
// This curve is intended for circuit operations defined over the KoalaBear field.
// In particular, it is used for multiset-hash constructions appearing in zkVM
// memory arguments.
//
// kb8: A curve over 𝔽p⁸ with
//
//	𝔽p: p=2130706433 = 2³¹-2²⁴+1
//	𝔽r: r=424804331891979973455971894938199991839487883914575852667663156896715214921
//	𝔽p²[u] = 𝔽p/u²-3
//	𝔽p⁴[v] = 𝔽p²/v²-u
//	𝔽p⁸[w] = 𝔽p⁴/w²-v
//	(E/𝔽p⁸): Y²=X³-3X+17w⁵
//	r ∣ #E(𝔽p⁸)
//
// # Warning
//
// This code has not been audited and is provided as-is. In particular, there
// is no security guarantee such as constant time implementation or side-channel
// attack resistance.
package kb8

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/kb8/internal/fptower"
)

// ID kb8 ID.
const ID = ecc.KB8

// aCurveCoeff and bCurveCoeff are the coefficients of Y²=X³+ax+b.
var aCurveCoeff, bCurveCoeff fptower.E8

// Generator and infinity point of G1.
var (
	g1Gen      G1Jac
	g1GenAff   G1Affine
	g1Infinity G1Jac
)

// xGen is only used by the generic mulBySeed helper. kb8 has no seed-based endomorphism,
// so keep it as the identity scalar.
var xGen big.Int

func init() {
	aCurveCoeff.C0.B0.A0.SetUint64(3)
	aCurveCoeff.Neg(&aCurveCoeff)

	bCurveCoeff.C1.B0.A1.SetUint64(17)

	g1Gen.X.C0.B0.A0.SetUint64(4)
	g1Gen.Y.SetString(
		"177975122", "773296979",
		"473899551", "417630813",
		"1724315640", "307114955",
		"459074134", "668770585",
	)
	g1Gen.Z.SetOne()
	g1GenAff.FromJacobian(&g1Gen)

	g1Infinity.X.SetOne()
	g1Infinity.Y.SetOne()

	xGen.SetInt64(1)
}

// Generators returns the generator of G1 in Jacobian and affine form.
func Generators() (g1Jac G1Jac, g1Aff G1Affine) {
	g1Jac = g1Gen
	g1Aff = g1GenAff
	return
}

// CurveCoefficients returns the coefficients of the curve equation.
func CurveCoefficients() (a, b fptower.E8) {
	return aCurveCoeff, bCurveCoeff
}
